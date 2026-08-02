package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMetaAngleKeysJumpToChatListEnds(t *testing.T) {
	app := NewApp()
	app.Chats = []Chat{{ID: "first"}, {ID: "middle"}, {ID: "last"}}
	app.SelectedIndex = 1
	model := NewModel(app, "client", "user")

	firstKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'<'}, Alt: true}
	if firstKey.String() != "alt+<" {
		t.Fatalf("unexpected Bubble Tea Meta-< encoding: %q", firstKey.String())
	}
	model, _ = model.handleNormalModeKey(firstKey)
	if model.app.SelectedIndex != 0 {
		t.Fatalf("Meta-< selected index %d, want 0", model.app.SelectedIndex)
	}

	lastKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'>'}, Alt: true}
	if lastKey.String() != "alt+>" {
		t.Fatalf("unexpected Bubble Tea Meta-> encoding: %q", lastKey.String())
	}
	model, _ = model.handleNormalModeKey(lastKey)
	if model.app.SelectedIndex != len(model.app.Chats)-1 {
		t.Fatalf("Meta-> selected index %d, want %d", model.app.SelectedIndex, len(model.app.Chats)-1)
	}
}

func TestFilteredNavigationVisitsEveryVisibleChatInOrder(t *testing.T) {
	app := NewApp()
	model := NewModel(app, "client", "user")
	model.latestChats = []Chat{
		{ID: "group-a", ChatType: "group"},
		{ID: "direct", ChatType: "oneOnOne"},
		{ID: "group-b", ChatType: "group"},
		{ID: "group-c", ChatType: "group"},
	}
	model.stableChatOrder = []string{"group-a", "direct", "group-b", "group-c"}
	app.ActiveChatFilter.ChatTypes["group"] = true
	model = model.rebuildChatList()

	want := []string{"group-b", "group-c", "group-a"}
	for index, wantID := range want {
		model, _ = model.handleNormalModeKey(filterTestKey('j'))
		selected := model.app.GetSelectedChat()
		if selected == nil || selected.ID != wantID {
			t.Fatalf("step %d selected %#v, want %s", index, selected, wantID)
		}
	}
}

func TestRebuildPreservesExplicitDashboardSelection(t *testing.T) {
	app := NewApp()
	app.SelectedIndex = -1
	model := NewModel(app, "client", "user")
	model.latestChats = []Chat{{ID: "first", ChatType: "group"}}
	model.stableChatOrder = []string{"first"}

	model = model.rebuildChatList()
	if model.app.SelectedIndex != -1 {
		t.Fatalf("background rebuild reactivated chat index %d", model.app.SelectedIndex)
	}
}

func TestMessageLoadCompletionUsesChatIDAfterListChanges(t *testing.T) {
	app := NewApp()
	app.Chats = []Chat{{ID: "visible"}}
	app.SelectedIndex = 0
	app.Messages = []Message{{ID: "visible-message"}}
	model := NewModel(app, "client", "user")

	model, _ = model.updateInternal(MsgMessagesLoaded{
		ChatID:   "formerly-at-index-zero",
		Messages: []Message{{ID: "background-message"}},
	})

	if got := model.app.CachedMessages["formerly-at-index-zero"]; len(got) != 1 || got[0].ID != "background-message" {
		t.Fatalf("response was not cached under immutable chat ID: %#v", got)
	}
	if len(model.app.Messages) != 1 || model.app.Messages[0].ID != "visible-message" {
		t.Fatalf("stale response replaced visible messages: %#v", model.app.Messages)
	}
}
