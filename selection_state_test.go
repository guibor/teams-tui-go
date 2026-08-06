package main

import (
	"strings"
	"testing"
)

func TestSelectedChatIdentitySurvivesVisibleListReordering(t *testing.T) {
	app := NewApp()
	app.SetChats([]Chat{{ID: "first"}, {ID: "second"}, {ID: "third"}})
	if !app.SetSelectedChatID("second") {
		t.Fatal("could not select second chat")
	}

	app.SetChats([]Chat{{ID: "third"}, {ID: "second"}, {ID: "first"}})
	selected := app.GetSelectedChat()
	if selected == nil || selected.ID != "second" || app.SelectedIndex != 1 {
		t.Fatalf("selection drifted after reorder: selected=%#v index=%d", selected, app.SelectedIndex)
	}
}

func TestSidebarRenderingRepairsStaleRowIndexFromSelectedIdentity(t *testing.T) {
	firstName := "First"
	secondName := "Second"
	app := NewApp()
	app.SetChats([]Chat{
		{ID: "first", CachedDisplayName: &firstName},
		{ID: "second", CachedDisplayName: &secondName},
	})
	app.SetSelectedChatID("second")
	app.SelectedIndex = 0
	app.SetMessages("second", []Message{{ID: "second-message", ChatID: "second"}}, "")
	model := NewModel(app, "client", "user")

	lines := strings.Split(stripANSI(model.renderChatList(50, 10)), "\n")
	if app.SelectedIndex != 1 {
		t.Fatalf("render retained stale row index %d, want 1", app.SelectedIndex)
	}
	for _, line := range lines {
		if strings.Contains(line, firstName) && strings.Contains(line, "›") {
			t.Fatalf("stale first row remained highlighted: %q", line)
		}
		if strings.Contains(line, secondName) && !strings.Contains(line, "›") {
			t.Fatalf("canonical second row was not highlighted: %q", line)
		}
	}
	if got := model.activeConversationTitle(); got != secondName {
		t.Fatalf("right-pane title = %q, want %q", got, secondName)
	}
}

func TestMismatchedMessageMetadataForcesTranscriptReload(t *testing.T) {
	app := NewApp()
	app.SetChats([]Chat{{ID: "first"}, {ID: "second"}})
	app.SetSelectedChatID("second")
	app.MessagesConversationID = "second"
	app.Messages = []Message{{ID: "wrong-message", ChatID: "first"}}
	model := NewModel(app, "client", "user")

	model, cmd := model.reconcileSelectedChatConversation()
	if cmd == nil {
		t.Fatal("mismatched transcript did not trigger a reload")
	}
	if model.app.MessagesConversationID != "second" || len(model.app.Messages) != 0 {
		t.Fatalf("mismatched transcript was not cleared: owner=%q messages=%#v", model.app.MessagesConversationID, model.app.Messages)
	}
	if !model.app.LoadingMessages {
		t.Fatal("replacement transcript was not marked loading")
	}
}

func TestRebuildMovesIdentityWhenFilterRemovesSelectedChat(t *testing.T) {
	app := NewApp()
	model := NewModel(app, "client", "user")
	model.latestChats = []Chat{
		{ID: "first", ChatType: "group"},
		{ID: "second", ChatType: "group"},
	}
	model.stableChatOrder = []string{"first", "second"}
	model = model.rebuildChatList()
	if !app.SetSelectedChatID("first") {
		t.Fatal("could not select first chat")
	}

	delete(model.chatCache, "first")
	model.latestChats = []Chat{{ID: "second", ChatType: "group"}}
	model.stableChatOrder = []string{"second"}
	model = model.rebuildChatList()

	selected := app.GetSelectedChat()
	if selected == nil || selected.ID != "second" || app.SelectedChatID != "second" {
		t.Fatalf("filter removal left stale selection: selected=%#v id=%q", selected, app.SelectedChatID)
	}
}

func TestSupersededMessageResponseCannotReplaceNewerTranscript(t *testing.T) {
	app := NewApp()
	app.SetChats([]Chat{{ID: "chat"}})
	app.SetSelectedChatID("chat")
	app.ActivateMessagesConversation("chat")
	model := NewModel(app, "client", "user")
	model.messageRequestGeneration["chat"] = 2

	model, _ = model.updateInternal(MsgMessagesLoaded{
		ChatID:     "chat",
		Generation: 1,
		Messages:   []Message{{ID: "stale"}},
	})
	if len(app.Messages) != 0 || len(app.CachedMessages["chat"]) != 0 {
		t.Fatalf("superseded response mutated state: messages=%#v cache=%#v", app.Messages, app.CachedMessages["chat"])
	}

	model, _ = model.updateInternal(MsgMessagesLoaded{
		ChatID:     "chat",
		Generation: 2,
		Messages:   []Message{{ID: "fresh"}},
	})
	if len(app.Messages) != 1 || app.Messages[0].ID != "fresh" {
		t.Fatalf("current response was not displayed: %#v", app.Messages)
	}
}
