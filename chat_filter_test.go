package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func filterTestString(value string) *string {
	return &value
}

func filterTestMessage(id, sender string) *Message {
	return &Message{
		ID:              id,
		CreatedDateTime: "2026-08-02T10:00:00Z",
		From: &MessageFrom{User: &MessageUser{
			DisplayName: filterTestString(sender),
		}},
	}
}

func filterTestKey(value rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{value}}
}

func TestTeamsChatURLUsesOpaqueGraphURL(t *testing.T) {
	chat := &Chat{WebURL: "  https://teams.microsoft.com/l/chat/example  "}
	if got := teamsChatURL(chat); got != "https://teams.microsoft.com/l/chat/example" {
		t.Fatalf("unexpected Teams chat URL: %q", got)
	}
	if got := teamsChatURL(nil); got != "" {
		t.Fatalf("nil chat should not have a Teams URL: %q", got)
	}
}

func TestNormalModeOpenUsesSelectedChatTeamsURL(t *testing.T) {
	app := NewApp()
	app.BrowserCommand = "true"
	app.Chats = []Chat{{
		ID:                "chat-1",
		WebURL:            "https://teams.microsoft.com/l/chat/example",
		CachedDisplayName: filterTestString("Example chat"),
	}}
	app.SelectedIndex = 0
	model := NewModel(app, "client", "user")
	_, cmd := model.handleNormalModeKey(filterTestKey('o'))
	if cmd == nil {
		t.Fatal("normal-mode o did not create an open-URL command")
	}
	if app.Status != "Opening chat in Teams..." {
		t.Fatalf("unexpected open status: %q", app.Status)
	}
}

func TestUnicodeChatIconsAreSemantic(t *testing.T) {
	app := NewApp()
	app.ChatIconTheme = "unicode"
	model := NewModel(app, "client", "user")

	tests := map[string]string{
		"oneOnOne": "@",
		"group":    "&",
		"meeting":  "◷",
		"channel":  "#",
	}
	for chatType, want := range tests {
		if got := model.chatTypeToIcon(chatType); got != want {
			t.Errorf("chat type %q: got %q, want %q", chatType, got, want)
		}
	}
}

func TestChatFilterCombinesReadTypeFavouriteAndText(t *testing.T) {
	app := NewApp()
	currentUser := "Me"
	app.CurrentUserName = &currentUser
	model := NewModel(app, "client", "user")
	model.favourites["direct"] = true

	direct := Chat{
		ID:                 "direct",
		ChatType:           "oneOnOne",
		CachedDisplayName:  filterTestString("Alice Example"),
		LastMessagePreview: filterTestMessage("message-1", "Alice Example"),
		Members: []ChatMember{{
			DisplayName: filterTestString("Alice Example"),
			Email:       filterTestString("alice@example.com"),
		}},
	}
	model.lastMsgID[direct.ID] = "message-1"

	filter := newChatListFilter()
	filter.ReadState = ChatReadUnread
	filter.ChatTypes["oneOnOne"] = true
	filter.FavouritesOnly = true
	filter.Query = "alice@example"
	if !model.chatMatchesFilter(direct, filter) {
		t.Fatal("expected direct chat to match all combined criteria")
	}

	filter.ReadState = ChatReadRead
	if model.chatMatchesFilter(direct, filter) {
		t.Fatal("unread chat matched read-only filter")
	}
	filter.ReadState = ChatReadUnread
	filter.ChatTypes = map[string]bool{"group": true}
	if model.chatMatchesFilter(direct, filter) {
		t.Fatal("direct chat matched group-only filter")
	}
}

func TestRebuildChatListRestoresChatsAfterFilterClears(t *testing.T) {
	app := NewApp()
	currentUser := "Me"
	app.CurrentUserName = &currentUser
	model := NewModel(app, "client", "user")

	unread := Chat{
		ID:                 "unread",
		ChatType:           "oneOnOne",
		CachedDisplayName:  filterTestString("Unread chat"),
		LastMessagePreview: filterTestMessage("message-1", "Someone else"),
	}
	read := Chat{
		ID:                 "read",
		ChatType:           "group",
		CachedDisplayName:  filterTestString("Read chat"),
		LastMessagePreview: filterTestMessage("message-2", "Someone else"),
	}
	model.latestChats = []Chat{unread, read}
	model.stableChatOrder = []string{unread.ID, read.ID}
	model.lastMsgID[unread.ID] = "message-1"
	model.lastMsgID[read.ID] = "message-2"
	model.lastReadMsgID[read.ID] = "message-2"

	filter := newChatListFilter()
	filter.ReadState = ChatReadUnread
	app.ActiveChatFilter = filter
	model = model.rebuildChatList()
	if len(app.Chats) != 1 || app.Chats[0].ID != unread.ID {
		t.Fatalf("unexpected unread-only list: %#v", app.Chats)
	}

	// Prove clearing the filter restores from the model cache, even between API refreshes.
	model.latestChats = nil
	app.ActiveChatFilter = newChatListFilter()
	model = model.rebuildChatList()
	if len(app.Chats) != 2 {
		t.Fatalf("clearing filter restored %d chats, want 2", len(app.Chats))
	}
}

func TestCloneChatListFilterDoesNotShareTypeMap(t *testing.T) {
	original := newChatListFilter()
	original.ChatTypes["group"] = true
	clone := cloneChatListFilter(original)
	delete(clone.ChatTypes, "group")
	if !original.ChatTypes["group"] {
		t.Fatal("editing draft filter mutated active filter")
	}
}

func TestChatFilterPopupAppliesAndCancelsDrafts(t *testing.T) {
	app := NewApp()
	currentUser := "Me"
	app.CurrentUserName = &currentUser
	model := NewModel(app, "client", "user")
	unread := Chat{
		ID:                 "unread",
		ChatType:           "oneOnOne",
		CachedDisplayName:  filterTestString("Unread chat"),
		LastMessagePreview: filterTestMessage("message-1", "Someone else"),
	}
	read := Chat{
		ID:                 "read",
		ChatType:           "group",
		CachedDisplayName:  filterTestString("Read chat"),
		LastMessagePreview: filterTestMessage("message-2", "Someone else"),
	}
	model.latestChats = []Chat{read, unread}
	model.stableChatOrder = []string{read.ID, unread.ID}
	model.lastMsgID[read.ID] = "message-2"
	model.lastMsgID[unread.ID] = "message-1"
	model.lastReadMsgID[read.ID] = "message-2"
	model = model.rebuildChatList()

	model, _ = model.handleNormalModeKey(filterTestKey('F'))
	if !app.ChatFilterPopupMode {
		t.Fatal("F did not open the chat filter popup")
	}
	model, _ = model.handleChatFilterPopupKey(filterTestKey('u'))
	model, _ = model.handleChatFilterPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if app.ActiveChatFilter.ReadState != ChatReadUnread {
		t.Fatalf("active read filter is %q, want unread", app.ActiveChatFilter.ReadState)
	}
	if len(app.Chats) != 1 || app.Chats[0].ID != unread.ID {
		t.Fatalf("unexpected filtered chats: %#v", app.Chats)
	}

	model, _ = model.handleNormalModeKey(filterTestKey('F'))
	model, _ = model.handleChatFilterPopupKey(filterTestKey('r'))
	model, _ = model.handleChatFilterPopupKey(tea.KeyMsg{Type: tea.KeyEsc})
	if app.ActiveChatFilter.ReadState != ChatReadUnread {
		t.Fatal("canceling the popup changed the active filter")
	}
}

func TestReadStateChangeImmediatelyReappliesUnreadFilter(t *testing.T) {
	app := NewApp()
	currentUser := "Me"
	app.CurrentUserName = &currentUser
	model := NewModel(app, "client", "user")
	first := Chat{
		ID:                 "first",
		ChatType:           "oneOnOne",
		CachedDisplayName:  filterTestString("First"),
		LastMessagePreview: filterTestMessage("message-1", "Someone else"),
	}
	second := Chat{
		ID:                 "second",
		ChatType:           "group",
		CachedDisplayName:  filterTestString("Second"),
		LastMessagePreview: filterTestMessage("message-2", "Someone else"),
	}
	model.latestChats = []Chat{first, second}
	model.stableChatOrder = []string{first.ID, second.ID}
	model.lastMsgID[first.ID] = "message-1"
	model.lastMsgID[second.ID] = "message-2"
	filter := newChatListFilter()
	filter.ReadState = ChatReadUnread
	app.ActiveChatFilter = filter
	model = model.rebuildChatList()

	model, _ = model.updateInternal(MsgChatReadStateChanged{ChatID: first.ID})
	if len(app.Chats) != 1 || app.Chats[0].ID != second.ID {
		t.Fatalf("read chat remained in unread filter: %#v", app.Chats)
	}
	if selected := app.GetSelectedChat(); selected == nil || selected.ID != second.ID {
		t.Fatalf("selection did not move to remaining unread chat: %#v", selected)
	}
}
