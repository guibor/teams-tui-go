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

func TestChatEndpointKeysUseVisibleChatList(t *testing.T) {
	app := NewApp()
	app.Chats = []Chat{{ID: "first"}, {ID: "middle"}, {ID: "last"}}
	app.SelectedIndex = 1
	model := NewModel(app, "client", "user")

	model, _ = model.handleNormalModeKey(filterTestKey('g'))
	if model.app.SelectedIndex != 1 || !model.pendingChatGoto {
		t.Fatalf("first g selected index %d with pending=%v, want index 1 and pending sequence", model.app.SelectedIndex, model.pendingChatGoto)
	}
	model, _ = model.handleNormalModeKey(filterTestKey('g'))
	if model.app.SelectedIndex != 0 || model.pendingChatGoto {
		t.Fatalf("gg selected index %d with pending=%v, want first index and completed sequence", model.app.SelectedIndex, model.pendingChatGoto)
	}

	model.app.SelectedIndex = 1
	model, _ = model.handleNormalModeKey(filterTestKey('G'))
	if model.app.SelectedIndex != 2 {
		t.Fatalf("G selected index %d, want last index 2", model.app.SelectedIndex)
	}

	model, _ = model.handleNormalModeKey(filterTestKey('h'))
	if model.app.SelectedIndex != 0 {
		t.Fatalf("h selected index %d, want first index 0", model.app.SelectedIndex)
	}

	model, _ = model.handleNormalModeKey(filterTestKey('l'))
	if model.app.SelectedIndex != 2 {
		t.Fatalf("l selected index %d, want last index 2", model.app.SelectedIndex)
	}
}

func TestChatEndpointKeysRespectActiveFilter(t *testing.T) {
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
	model.app.SelectedIndex = 1

	model, _ = model.handleNormalModeKey(filterTestKey('l'))
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != "group-c" {
		t.Fatalf("l selected %#v, want last visible chat group-c", selected)
	}
	model, _ = model.handleNormalModeKey(filterTestKey('h'))
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != "group-a" {
		t.Fatalf("h selected %#v, want first visible chat group-a", selected)
	}
	model, _ = model.handleNormalModeKey(filterTestKey('G'))
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != "group-c" {
		t.Fatalf("G selected %#v, want last visible chat group-c", selected)
	}
	model, _ = model.handleNormalModeKey(filterTestKey('g'))
	model, _ = model.handleNormalModeKey(filterTestKey('g'))
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != "group-a" {
		t.Fatalf("gg selected %#v, want first visible chat group-a", selected)
	}
}

func TestPlainAngleAndUppercaseKeysJumpMessagePaneEnds(t *testing.T) {
	for _, test := range []struct {
		key          rune
		wantOffset   int
		wantSnapDown bool
	}{
		{key: '<', wantOffset: 0, wantSnapDown: false},
		{key: 'H', wantOffset: 0, wantSnapDown: false},
		{key: '>', wantOffset: 120, wantSnapDown: true},
		{key: 'L', wantOffset: 120, wantSnapDown: true},
	} {
		app := NewApp()
		app.Chats = []Chat{{ID: "chat"}}
		app.SelectedIndex = 0
		app.ScrollOffset = 60
		app.MaxScroll = 120
		app.SnapToBottom = false
		model := NewModel(app, "client", "user")

		model, _ = model.handleNormalModeKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{test.key}})
		if model.app.ScrollOffset != test.wantOffset || model.app.SnapToBottom != test.wantSnapDown {
			t.Fatalf("%q produced offset=%d snap=%v, want offset=%d snap=%v",
				test.key, model.app.ScrollOffset, model.app.SnapToBottom, test.wantOffset, test.wantSnapDown)
		}
	}
}

func TestMessageSelectionEndKeysRespectNewestFirstStorage(t *testing.T) {
	app := NewApp()
	app.Messages = []Message{{ID: "newest"}, {ID: "middle"}, {ID: "oldest"}}
	app.MessageSelectionMode = true
	app.MessageSelectedIndex = 1
	model := NewModel(app, "client", "user")

	model, _ = model.handleMessageSelectionModeKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}})
	if model.app.MessageSelectedIndex != 2 {
		t.Fatalf("H selected index %d, want oldest index 2", model.app.MessageSelectedIndex)
	}
	model, _ = model.handleMessageSelectionModeKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}})
	if model.app.MessageSelectedIndex != 0 {
		t.Fatalf("L selected index %d, want newest index 0", model.app.MessageSelectedIndex)
	}
}

func TestMessagePopupEndKeysRespectNewestFirstStorage(t *testing.T) {
	app := NewApp()
	app.Messages = []Message{{ID: "newest"}, {ID: "middle"}, {ID: "oldest"}}
	app.MessageSelectedIndex = 1
	model := NewModel(app, "client", "user")

	model, _ = model.handleMessagePopupKey(filterTestKey('H'))
	if model.app.MessageSelectedIndex != 2 {
		t.Fatalf("H selected index %d, want oldest index 2", model.app.MessageSelectedIndex)
	}
	model, _ = model.handleMessagePopupKey(filterTestKey('L'))
	if model.app.MessageSelectedIndex != 0 {
		t.Fatalf("L selected index %d, want newest index 0", model.app.MessageSelectedIndex)
	}
}

func TestMetaNPNavigateChats(t *testing.T) {
	app := NewApp()
	app.Chats = []Chat{{ID: "first"}, {ID: "middle"}, {ID: "last"}}
	app.SelectedIndex = 1
	model := NewModel(app, "client", "user")

	nextKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}, Alt: true}
	if nextKey.String() != "alt+n" {
		t.Fatalf("unexpected Bubble Tea Meta-n encoding: %q", nextKey.String())
	}
	model, _ = model.handleNormalModeKey(nextKey)
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != "last" {
		t.Fatalf("Meta-n selected %#v, want last", selected)
	}

	previousKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}, Alt: true}
	if previousKey.String() != "alt+p" {
		t.Fatalf("unexpected Bubble Tea Meta-p encoding: %q", previousKey.String())
	}
	model, _ = model.handleNormalModeKey(previousKey)
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != "middle" {
		t.Fatalf("Meta-p selected %#v, want middle", selected)
	}
}

func TestMetaNNavigatesFilteredChatListWithoutSkipping(t *testing.T) {
	app := NewApp()
	model := NewModel(app, "client", "user")
	model.latestChats = []Chat{
		{ID: "group-a", ChatType: "group"},
		{ID: "direct", ChatType: "oneOnOne"},
		{ID: "group-b", ChatType: "group"},
	}
	model.stableChatOrder = []string{"group-a", "direct", "group-b"}
	app.ActiveChatFilter.ChatTypes["group"] = true
	model = model.rebuildChatList()

	metaNext := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}, Alt: true}
	model, _ = model.handleNormalModeKey(metaNext)
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != "group-b" {
		t.Fatalf("Meta-n selected %#v, want group-b", selected)
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
