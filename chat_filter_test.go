package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	if app.Status != "Opening chat in browser..." {
		t.Fatalf("unexpected open status: %q", app.Status)
	}
}

func TestNormalModeUpperOUsesTeamsDesktopLink(t *testing.T) {
	app := NewApp()
	app.TeamsAppCommand = "true"
	app.Chats = []Chat{{ID: "chat-1", WebURL: "https://teams.microsoft.com/l/chat/example"}}
	app.SelectedIndex = 0
	model := NewModel(app, "client", "user")

	_, cmd := model.handleNormalModeKey(filterTestKey('O'))
	if cmd == nil {
		t.Fatal("normal-mode O did not create a Teams-app command")
	}
	if app.Status != "Opening chat in Teams..." {
		t.Fatalf("unexpected Teams open status: %q", app.Status)
	}
}

func TestTeamsDesktopURLUsesOfficialScheme(t *testing.T) {
	got := teamsDesktopURL("https://teams.microsoft.com/l/chat/example?tenantId=abc")
	want := "msteams://teams.microsoft.com/l/chat/example?tenantId=abc"
	if got != want {
		t.Fatalf("desktop URL = %q, want %q", got, want)
	}
	if got := teamsDesktopURL("https://example.com/l/chat/example"); got != "" {
		t.Fatalf("non-Teams URL converted to %q", got)
	}
}

func TestTeamsWebURLBypassesLauncherWithoutChangingOpaqueTarget(t *testing.T) {
	got := teamsWebURL(" https://teams.microsoft.com/l/chat/19%3Ameeting_NjQ%40thread.v2/conversations?tenantId=abc&groupId=def ")
	want := "https://teams.microsoft.com/#/l/chat/19%3Ameeting_NjQ%40thread.v2/conversations?tenantId=abc&groupId=def"
	if got != want {
		t.Fatalf("web URL = %q, want %q", got, want)
	}
	if got := teamsWebURL("https://teams.cloud.microsoft/l/chat/example/conversations"); got != "https://teams.cloud.microsoft/#/l/chat/example/conversations" {
		t.Fatalf("new Teams host web URL = %q", got)
	}
	if got := teamsWebURL("https://teams.microsoft.com/#/l/chat/example/conversations"); got != "https://teams.microsoft.com/#/l/chat/example/conversations" {
		t.Fatalf("existing web route changed to %q", got)
	}
	if got := teamsWebURL("https://example.com/l/chat/example"); got != "https://example.com/l/chat/example" {
		t.Fatalf("non-Teams URL changed to %q", got)
	}
}

func TestThreadActionMenuCapturesSelectedChat(t *testing.T) {
	app := NewApp()
	title := "Capture me"
	app.Chats = []Chat{{ID: "chat-1", CachedDisplayName: &title}}
	app.SelectedIndex = 0
	app.ThreadCaptureFile = filepath.Join(t.TempDir(), "threads.md")
	model := NewModel(app, "client", "user")

	model, _ = model.handleNormalModeKey(filterTestKey('a'))
	if !app.ThreadActionPopupMode {
		t.Fatal("a did not open the thread action menu")
	}
	model, cmd := model.handleThreadActionPopupKey(filterTestKey('a'))
	if cmd == nil || app.ThreadActionPopupMode {
		t.Fatal("capture action did not close the menu and return a command")
	}
	rawMsg := cmd()
	msg, ok := rawMsg.(MsgThreadCaptured)
	if !ok {
		t.Fatalf("capture command returned %T", rawMsg)
	}
	model, _ = model.updateInternal(msg)
	if _, err := os.Stat(app.ThreadCaptureFile); err != nil {
		t.Fatalf("capture action did not create file: %v", err)
	}
}

func TestThreadActionMenuCapturesSelectedChatAsOrg(t *testing.T) {
	app := NewApp()
	title := "Capture me in Org"
	app.Chats = []Chat{{ID: "chat-org", CachedDisplayName: &title}}
	app.SelectedIndex = 0
	app.ThreadCaptureFormat = ThreadCaptureOrg
	app.ThreadCaptureOrgFile = filepath.Join(t.TempDir(), "threads.org")
	model := NewModel(app, "client", "user")

	model, _ = model.handleNormalModeKey(filterTestKey('a'))
	model, cmd := model.handleThreadActionPopupKey(filterTestKey('a'))
	if cmd == nil || app.ThreadActionPopupMode {
		t.Fatal("Org capture action did not close the menu and return a command")
	}
	rawMsg := cmd()
	msg, ok := rawMsg.(MsgThreadCaptured)
	if !ok {
		t.Fatalf("Org capture command returned %T", rawMsg)
	}
	model, _ = model.updateInternal(msg)
	data, err := os.ReadFile(app.ThreadCaptureOrgFile)
	if err != nil {
		t.Fatalf("Org capture action did not create file: %v", err)
	}
	if !strings.Contains(string(data), "** TODO Teams: Capture me in Org") {
		t.Fatalf("Org capture did not use Org format:\n%s", data)
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

	model, _ = model.handleNormalModeKey(filterTestKey('v'))
	if !app.ChatFilterPopupMode {
		t.Fatal("v did not open the chat filter popup")
	}
	model, _ = model.handleChatFilterPopupKey(filterTestKey('u'))
	model, _ = model.handleChatFilterPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if app.ActiveChatFilter.ReadState != ChatReadUnread {
		t.Fatalf("active read filter is %q, want unread", app.ActiveChatFilter.ReadState)
	}
	if len(app.Chats) != 1 || app.Chats[0].ID != unread.ID {
		t.Fatalf("unexpected filtered chats: %#v", app.Chats)
	}

	model, _ = model.handleNormalModeKey(filterTestKey('v'))
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

func TestReadStateChangeReappliesStructuredUnreadQuery(t *testing.T) {
	app := NewApp()
	currentUser := "Me"
	app.CurrentUserName = &currentUser
	model := NewModel(app, "client", "user")
	first := Chat{ID: "first", LastMessagePreview: filterTestMessage("message-1", "Someone else")}
	second := Chat{ID: "second", LastMessagePreview: filterTestMessage("message-2", "Someone else")}
	model.latestChats = []Chat{first, second}
	model.stableChatOrder = []string{first.ID, second.ID}
	model.lastMsgID[first.ID] = "message-1"
	model.lastMsgID[second.ID] = "message-2"
	app.ActiveChatFilter.Query = "is:unread"
	model = model.rebuildChatList()
	app.SetSelectedChatID(first.ID)
	app.SetMessages(first.ID, []Message{{ID: "first-transcript", ChatID: first.ID}}, "")
	app.ChatMessagesLoadedOnce[second.ID] = true
	app.CachedMessages[second.ID] = []Message{{ID: "second-transcript", ChatID: second.ID}}

	model, cmd := model.updateInternal(MsgChatReadStateChanged{ChatID: first.ID})
	if cmd != nil {
		t.Fatal("cached query-filter replacement unexpectedly requested a network load")
	}
	if len(app.Chats) != 1 || app.Chats[0].ID != second.ID {
		t.Fatalf("structured is:unread query retained the read chat: %#v", app.Chats)
	}
	if selected := app.GetSelectedChat(); selected == nil || selected.ID != second.ID {
		t.Fatalf("structured query selected %#v, want %q", selected, second.ID)
	}
	if !app.MessagesBelongTo(second.ID) || len(app.Messages) != 1 || app.Messages[0].ID != "second-transcript" {
		t.Fatalf("structured query transcript does not match selection: owner=%q messages=%#v", app.MessagesConversationID, app.Messages)
	}
}

func TestReadStateChangeForBackgroundChatDoesNotResetSelectedTranscript(t *testing.T) {
	app := NewApp()
	currentUser := "Me"
	app.CurrentUserName = &currentUser
	model := NewModel(app, "client", "user")
	first := Chat{ID: "first", LastMessagePreview: filterTestMessage("message-1", "Someone else")}
	second := Chat{ID: "second", LastMessagePreview: filterTestMessage("message-2", "Someone else")}
	model.latestChats = []Chat{first, second}
	model.stableChatOrder = []string{first.ID, second.ID}
	model.lastMsgID[first.ID] = "message-1"
	model.lastMsgID[second.ID] = "message-2"
	app.ActiveChatFilter.ReadState = ChatReadUnread
	model = model.rebuildChatList()
	app.SetSelectedChatIndex(1)
	app.SetMessages(second.ID, []Message{{ID: "selected-transcript", ChatID: second.ID}}, "")

	model, cmd := model.updateInternal(MsgChatReadStateChanged{ChatID: first.ID})
	if cmd != nil {
		t.Fatal("background read-state update unexpectedly reloaded the selected chat")
	}
	if selected := app.GetSelectedChat(); selected == nil || selected.ID != second.ID {
		t.Fatalf("background update changed selection: %#v", selected)
	}
	if len(app.Messages) != 1 || app.Messages[0].ID != "selected-transcript" {
		t.Fatalf("background update reset the selected transcript: %#v", app.Messages)
	}
	if len(app.Chats) != 1 || app.Chats[0].ID != second.ID {
		t.Fatalf("read chat remained in unread filter: %#v", app.Chats)
	}
}

func TestReadStateChangeLoadsReplacementTranscriptByChatID(t *testing.T) {
	app := NewApp()
	currentUser := "Me"
	app.CurrentUserName = &currentUser
	model := NewModel(app, "client", "user")
	first := Chat{ID: "first", LastMessagePreview: filterTestMessage("message-1", "Someone else")}
	second := Chat{ID: "second", LastMessagePreview: filterTestMessage("message-2", "Someone else")}
	model.latestChats = []Chat{first, second}
	model.stableChatOrder = []string{first.ID, second.ID}
	model.lastMsgID[first.ID] = "message-1"
	model.lastMsgID[second.ID] = "message-2"
	app.ActiveChatFilter.ReadState = ChatReadUnread
	model = model.rebuildChatList()
	app.SelectedIndex = 0
	app.Messages = []Message{{ID: "first-transcript"}}
	app.ChatMessagesLoadedOnce[second.ID] = true
	app.CachedMessages[second.ID] = []Message{{ID: "second-transcript"}}

	model, cmd := model.updateInternal(MsgChatReadStateChanged{ChatID: first.ID})
	if cmd != nil {
		t.Fatal("cached replacement transcript unexpectedly requested a network load")
	}
	if selected := app.GetSelectedChat(); selected == nil || selected.ID != second.ID {
		t.Fatalf("wrong replacement selected: %#v", selected)
	}
	if len(app.Messages) != 1 || app.Messages[0].ID != "second-transcript" {
		t.Fatalf("replacement transcript does not match selected chat: %#v", app.Messages)
	}
}

func TestChatRefreshClearsTranscriptWhenFilteredSelectionDisappears(t *testing.T) {
	app := NewApp()
	currentUser := "Me"
	app.CurrentUserName = &currentUser
	first := Chat{
		ID:                 "first",
		CachedDisplayName:  filterTestString("First"),
		LastMessagePreview: filterTestMessage("first-message", "Someone else"),
	}
	second := Chat{
		ID:                 "second",
		CachedDisplayName:  filterTestString("Second"),
		LastMessagePreview: filterTestMessage("second-message", "Someone else"),
	}
	app.Chats = []Chat{first, second}
	app.SelectedIndex = 0
	app.ActiveChatFilter.ReadState = ChatReadUnread
	app.MessagesConversationID = first.ID
	app.Messages = []Message{{ID: "first-transcript"}}

	model := NewModel(app, "client", "user")
	model.latestChats = []Chat{first, second}
	model.stableChatOrder = []string{first.ID, second.ID}
	model.lastMsgID[first.ID] = "first-message"
	model.lastMsgID[second.ID] = "second-message"
	model.lastReadMsgID[first.ID] = "first-message"

	model, cmd := model.updateInternal(MsgChatsLoaded{Chats: []Chat{first, second}})
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != second.ID {
		t.Fatalf("refresh selected %#v, want replacement chat %q", selected, second.ID)
	}
	if model.app.MessagesConversationID != second.ID {
		t.Fatalf("transcript owner is %q, want %q", model.app.MessagesConversationID, second.ID)
	}
	if len(model.app.Messages) != 0 {
		t.Fatalf("old transcript remained under replacement selection: %#v", model.app.Messages)
	}
	if !model.app.LoadingMessages || cmd == nil {
		t.Fatalf("replacement transcript was not requested: loading=%v cmd=%v", model.app.LoadingMessages, cmd != nil)
	}
}

func TestMessageLoadForNewOwnerReplacesInsteadOfMergingTranscript(t *testing.T) {
	app := NewApp()
	app.Chats = []Chat{{ID: "second"}}
	app.SelectedIndex = 0
	app.MessagesConversationID = "first"
	app.Messages = []Message{{ID: "first-transcript"}}
	model := NewModel(app, "client", "user")

	model, _ = model.updateInternal(MsgMessagesLoaded{
		ChatID:   "second",
		Messages: []Message{{ID: "second-transcript"}},
	})

	if model.app.MessagesConversationID != "second" {
		t.Fatalf("transcript owner is %q, want second", model.app.MessagesConversationID)
	}
	if len(model.app.Messages) != 1 || model.app.Messages[0].ID != "second-transcript" {
		t.Fatalf("cross-chat transcripts were merged: %#v", model.app.Messages)
	}
}

func TestEmptyMessageLoadStillClaimsNewTranscriptOwner(t *testing.T) {
	app := NewApp()
	app.Chats = []Chat{{ID: "empty-chat"}}
	app.SelectedIndex = 0
	app.MessagesConversationID = "previous-chat"
	app.Messages = []Message{{ID: "previous-transcript"}}
	app.SetLoadingMessages(true)
	model := NewModel(app, "client", "user")

	model, _ = model.updateInternal(MsgMessagesLoaded{ChatID: "empty-chat"})

	if model.app.MessagesConversationID != "empty-chat" || len(model.app.Messages) != 0 {
		t.Fatalf("empty chat retained prior transcript state: owner=%q messages=%#v",
			model.app.MessagesConversationID, model.app.Messages)
	}
	if model.app.LoadingMessages {
		t.Fatal("empty chat remained stuck in loading state")
	}
}

func TestChatRefreshDoesNotReplaceActiveChannelTranscript(t *testing.T) {
	app := NewApp()
	chat := Chat{ID: "chat", LastMessagePreview: filterTestMessage("chat-message", "Someone else")}
	app.Chats = []Chat{chat}
	app.SelectedIndex = 0
	app.MessagesConversationID = "channel"
	app.Messages = []Message{{ID: "channel-transcript"}}
	model := NewModel(app, "client", "user")
	model.channelSelectedIndex = 0
	model.latestChats = []Chat{chat}
	model.stableChatOrder = []string{chat.ID}
	model.lastMsgID[chat.ID] = "chat-message"

	model, _ = model.updateInternal(MsgChatsLoaded{Chats: []Chat{chat}})

	if model.app.MessagesConversationID != "channel" || len(model.app.Messages) != 1 ||
		model.app.Messages[0].ID != "channel-transcript" {
		t.Fatalf("chat refresh replaced active channel transcript: owner=%q messages=%#v",
			model.app.MessagesConversationID, model.app.Messages)
	}
}

func TestNormalModeSOpensChatSearchAndCComposes(t *testing.T) {
	app := NewApp()
	model := NewModel(app, "client", "user")

	model, _ = model.handleNormalModeKey(filterTestKey('s'))
	if !app.UserSearchPopupMode || !app.UserSearchMode {
		t.Fatal("s did not open chat search")
	}

	app.UserSearchPopupMode = false
	app.UserSearchMode = false
	app.Chats = []Chat{{ID: "chat"}}
	app.SelectedIndex = 0
	model, _ = model.handleNormalModeKey(filterTestKey('c'))
	if !app.InputMode || app.UserSearchPopupMode || app.UserSearchMode {
		t.Fatal("c did not open the current chat composer")
	}
}

func TestNormalModeDTogglesAndPersistsChatDates(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	InitConfig()
	app := NewApp()
	model := NewModel(app, "client", "user")

	model, _ = model.handleNormalModeKey(filterTestKey('D'))
	if !app.ShowChatDates {
		t.Fatal("D did not enable chat dates")
	}
	cfg := LoadConfig()
	if cfg == nil || cfg.ShowChatDates == nil || !*cfg.ShowChatDates {
		t.Fatalf("D did not persist show_chat_dates: %#v", cfg)
	}

	model, _ = model.handleNormalModeKey(filterTestKey('D'))
	if app.ShowChatDates {
		t.Fatal("second D did not hide chat dates")
	}
}

func TestChatLastMessageTimestampUsesLocalDateTimeAndYear(t *testing.T) {
	location := time.FixedZone("IDT", 3*60*60)
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, location)
	chat := Chat{LastMessagePreview: &Message{CreatedDateTime: "2026-08-02T22:30:00Z"}}
	if got := chatLastMessageTimestamp(chat, now); got != "Aug 03 01:30" {
		t.Fatalf("local chat timestamp = %q, want Aug 03 01:30", got)
	}
	chat.LastMessagePreview.CreatedDateTime = "2025-12-31T20:00:00Z"
	if got := chatLastMessageTimestamp(chat, now); got != "2025-12-31 23:00" {
		t.Fatalf("prior-year chat timestamp = %q, want 2025-12-31 23:00", got)
	}
}

func TestChatBookmarkPrefixAppliesUnreadAndInboxPresets(t *testing.T) {
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

	model, _ = model.handleNormalModeKey(filterTestKey('b'))
	if !app.ChatBookmarkPopupMode {
		t.Fatal("b did not open bookmark presets")
	}
	model, _ = model.handleChatBookmarkPopupKey(filterTestKey('u'))
	if app.ActiveChatFilter.ReadState != ChatReadUnread {
		t.Fatalf("bu read filter is %q, want unread", app.ActiveChatFilter.ReadState)
	}
	if len(app.Chats) != 1 || app.Chats[0].ID != unread.ID {
		t.Fatalf("bu showed unexpected chats: %#v", app.Chats)
	}

	model, _ = model.handleNormalModeKey(filterTestKey('b'))
	model, _ = model.handleChatBookmarkPopupKey(filterTestKey('i'))
	if chatFilterIsActive(app.ActiveChatFilter) {
		t.Fatalf("bi left an active filter: %#v", app.ActiveChatFilter)
	}
	if len(app.Chats) != 2 {
		t.Fatalf("bi showed %d chats, want 2", len(app.Chats))
	}
}

func TestUpperUReplacesEveryActiveFilterWithUnreadOnly(t *testing.T) {
	app := NewApp()
	currentUser := "Me"
	app.CurrentUserName = &currentUser
	model := NewModel(app, "client", "user")
	unread := Chat{ID: "unread", ChatType: "oneOnOne", LastMessagePreview: filterTestMessage("unread-last", "Other")}
	read := Chat{ID: "read", ChatType: "group", LastMessagePreview: filterTestMessage("read-last", "Other")}
	model.latestChats = []Chat{read, unread}
	model.stableChatOrder = []string{read.ID, unread.ID}
	model.lastMsgID[read.ID] = "read-last"
	model.lastReadMsgID[read.ID] = "read-last"
	model.lastMsgID[unread.ID] = "unread-last"
	app.ActiveChatFilter = ChatListFilter{
		Query:          "something else",
		ReadState:      ChatReadRead,
		FavouritesOnly: true,
		TodayOnly:      true,
		ChatTypes:      map[string]bool{"group": true},
	}
	model = model.rebuildChatList()

	model, _ = model.handleNormalModeKey(filterTestKey('U'))
	filter := app.ActiveChatFilter
	if filter.ReadState != ChatReadUnread || filter.Query != "" || filter.FavouritesOnly || filter.TodayOnly || len(filter.ChatTypes) != 0 {
		t.Fatalf("U did not replace all filter criteria: %#v", filter)
	}
	if len(app.Chats) != 1 || app.Chats[0].ID != unread.ID {
		t.Fatalf("U showed unexpected chats: %#v", app.Chats)
	}
}

func TestFilteredChatListDeduplicatesCorruptStableOrder(t *testing.T) {
	app := NewApp()
	currentUser := "Me"
	app.CurrentUserName = &currentUser
	model := NewModel(app, "client", "user")
	first := Chat{ID: "first", ChatType: "group", LastMessagePreview: filterTestMessage("first-last", "Other")}
	second := Chat{ID: "second", ChatType: "group", LastMessagePreview: filterTestMessage("second-last", "Other")}
	model.latestChats = []Chat{first, first, second, second}
	model.stableChatOrder = []string{"first", "first", "second", "first", "second"}
	model.lastMsgID[first.ID] = "first-last"
	model.lastMsgID[second.ID] = "second-last"
	app.ActiveChatFilter.ReadState = ChatReadUnread

	model = model.rebuildChatList()
	if len(model.stableChatOrder) != 2 || model.stableChatOrder[0] != "first" || model.stableChatOrder[1] != "second" {
		t.Fatalf("stable order was not normalized: %#v", model.stableChatOrder)
	}
	if len(app.Chats) != 2 || app.Chats[0].ID != "first" || app.Chats[1].ID != "second" {
		t.Fatalf("filtered list contains duplicates: %#v", app.Chats)
	}
}

func TestMergeAndPromoteKeepStableOrderUnique(t *testing.T) {
	app := NewApp()
	model := NewModel(app, "client", "user")
	model.stableChatOrder = []string{"existing", "existing"}
	model = model.mergeChats([]Chat{{ID: "new"}, {ID: "new"}, {ID: "existing"}})
	if len(model.stableChatOrder) != 2 || model.stableChatOrder[0] != "existing" || model.stableChatOrder[1] != "new" {
		t.Fatalf("merge produced duplicate IDs: %#v", model.stableChatOrder)
	}

	model.stableChatOrder = []string{"existing", "new", "new", "existing"}
	model.promoteChat("new")
	if len(model.stableChatOrder) != 2 || model.stableChatOrder[0] != "new" || model.stableChatOrder[1] != "existing" {
		t.Fatalf("promotion produced duplicate IDs: %#v", model.stableChatOrder)
	}
}

func TestTodayBookmarkUsesLocalCalendarDay(t *testing.T) {
	location := time.FixedZone("IDT", 3*60*60)
	day := time.Date(2026, 8, 2, 10, 0, 0, 0, location)
	previousUTCDate := Chat{LastMessagePreview: &Message{CreatedDateTime: "2026-08-01T22:30:00Z"}}
	previousLocalDate := Chat{LastMessagePreview: &Message{CreatedDateTime: "2026-08-01T19:30:00Z"}}

	if !chatHasActivityOn(previousUTCDate, day) {
		t.Fatal("activity after local midnight did not match today's bookmark")
	}
	if chatHasActivityOn(previousLocalDate, day) {
		t.Fatal("activity before local midnight matched today's bookmark")
	}
}

func TestFilterReplacementAtSameIndexLoadsReplacementChat(t *testing.T) {
	app := NewApp()
	currentUser := "Me"
	app.CurrentUserName = &currentUser
	model := NewModel(app, "client", "user")
	read := Chat{ID: "read", ChatType: "group", LastMessagePreview: filterTestMessage("read-last", "Other")}
	unread := Chat{ID: "unread", ChatType: "group", LastMessagePreview: filterTestMessage("unread-last", "Other")}
	model.latestChats = []Chat{read, unread}
	model.stableChatOrder = []string{read.ID, unread.ID}
	model.lastMsgID[read.ID] = "read-last"
	model.lastReadMsgID[read.ID] = "read-last"
	model.lastMsgID[unread.ID] = "unread-last"
	model = model.rebuildChatList()
	app.Messages = []Message{{ID: "read-message"}}
	app.ChatMessagesLoadedOnce[unread.ID] = true
	app.CachedMessages[unread.ID] = []Message{{ID: "unread-message"}}

	app.DraftChatFilter = newChatListFilter()
	app.DraftChatFilter.ReadState = ChatReadUnread
	model, _ = model.applyChatFilter()

	if app.SelectedIndex != 0 {
		t.Fatalf("replacement selected index %d, want 0", app.SelectedIndex)
	}
	if selected := app.GetSelectedChat(); selected == nil || selected.ID != unread.ID {
		t.Fatalf("selected wrong replacement chat: %#v", selected)
	}
	if len(app.Messages) != 1 || app.Messages[0].ID != "unread-message" {
		t.Fatalf("same-index replacement retained stale messages: %#v", app.Messages)
	}
}
