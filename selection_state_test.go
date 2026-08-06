package main

import "testing"

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
