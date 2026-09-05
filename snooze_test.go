package main

import (
	"strings"
	"testing"
	"time"
)

func TestSnoozeChoicesUseConfiguredWorkday(t *testing.T) {
	app := NewApp()
	app.WorkdayStart = "07:00"
	app.WorkdayEnd = "18:00"
	model := NewModel(app, "client", "user")
	location := time.FixedZone("local", 2*60*60)

	beforeEnd := time.Date(2026, 9, 4, 16, 0, 0, 0, location)
	if got := snoozeChoices()[2].Until(model, beforeEnd); got.Hour() != 18 || got.Day() != 4 {
		t.Fatalf("end-of-day wake = %v", got)
	}
	afterEnd := time.Date(2026, 9, 4, 20, 0, 0, 0, location)
	if got := snoozeChoices()[2].Until(model, afterEnd); got.Hour() != 7 || got.Day() != 5 {
		t.Fatalf("after-hours wake = %v", got)
	}
	if got := snoozeChoices()[3].Until(model, beforeEnd); got.Hour() != 7 || got.Day() != 5 {
		t.Fatalf("tomorrow wake = %v", got)
	}
}

func TestQuickSnoozeHidesChatAndAdvancesWithoutChangingReadState(t *testing.T) {
	app := NewApp()
	first := Chat{ID: "first", LastMessagePreview: filterTestMessage("first-msg", "Other")}
	second := Chat{ID: "second", LastMessagePreview: filterTestMessage("second-msg", "Other")}
	app.SetChats([]Chat{first, second})
	app.SetSelectedChatID(first.ID)
	app.MessagesConversationID = first.ID
	model := NewModel(app, "client", "user")
	model.latestChats = []Chat{first, second}
	model.stableChatOrder = []string{first.ID, second.ID}
	model.lastMsgID[first.ID] = "first-msg"
	model.lastReadMsgID[first.ID] = "older"

	model, _ = model.quickSnooze()
	if len(app.Chats) != 1 || app.Chats[0].ID != second.ID || app.SelectedChatID != second.ID {
		t.Fatalf("snooze did not hide and advance: chats=%#v selected=%q", app.Chats, app.SelectedChatID)
	}
	if model.lastReadMsgID[first.ID] != "older" {
		t.Fatal("snooze changed read state")
	}
}

func TestSnoozedBookmarkShowsOnlySnoozedChats(t *testing.T) {
	model := NewModel(NewApp(), "client", "user")
	model.snoozed["sleeping"] = time.Now().Add(time.Hour)
	filter := newChatListFilter()
	filter.SnoozedOnly = true
	if !model.chatMatchesFilter(Chat{ID: "sleeping"}, filter) {
		t.Fatal("snoozed chat missing from bs")
	}
	if model.chatMatchesFilter(Chat{ID: "awake"}, filter) {
		t.Fatal("awake chat appeared in bs")
	}
}

func TestResnoozeRetainsChatWhenItRemainsInCurrentView(t *testing.T) {
	app := NewApp()
	first := Chat{ID: "first", LastMessagePreview: filterTestMessage("first-msg", "Other")}
	second := Chat{ID: "second", LastMessagePreview: filterTestMessage("second-msg", "Other")}
	app.SetChats([]Chat{first, second})
	app.SetSelectedChatID(first.ID)
	app.MessagesConversationID = first.ID
	filter := newChatListFilter()
	filter.SnoozedOnly = true
	app.ActiveChatFilter = filter
	model := NewModel(app, "client", "user")
	model.latestChats = []Chat{first, second}
	model.stableChatOrder = []string{first.ID, second.ID}
	model.snoozed[first.ID] = time.Now().Add(time.Hour)
	model.snoozed[second.ID] = time.Now().Add(time.Hour)
	model = model.rebuildChatList()
	app.SetSelectedChatID(first.ID)

	model, _ = model.applySnooze(time.Now().Add(3 * time.Hour))
	if app.SelectedChatID != first.ID || app.MessagesConversationID != first.ID {
		t.Fatalf("resnooze changed current chat: selected=%q transcript=%q", app.SelectedChatID, app.MessagesConversationID)
	}
}

func TestIncomingMessageWakesSnoozedChat(t *testing.T) {
	user := "Me"
	old := filterTestMessage("old", "Other")
	newest := filterTestMessage("new", "Other")
	newest.CreatedDateTime = "2026-08-02T10:05:00Z"
	chat := Chat{ID: "chat-1", LastMessagePreview: old}
	app := NewApp()
	app.CurrentUserName = &user
	app.SetChats([]Chat{chat})
	model := NewModel(app, "client", "user")
	model.latestChats = []Chat{chat}
	model.stableChatOrder = []string{chat.ID}
	model.lastMsgID[chat.ID] = old.ID
	model.lastMsgTime[chat.ID], _ = time.Parse(time.RFC3339, old.CreatedDateTime)
	model.snoozed[chat.ID] = time.Now().Add(3 * time.Hour)

	chat.LastMessagePreview = newest
	model, _ = model.updateInternal(MsgChatsLoaded{Chats: []Chat{chat}})
	if model.chatSnoozed(chat.ID, time.Now()) {
		t.Fatal("incoming message did not wake snoozed chat")
	}
}

func TestSnoozedChatsPersistActiveDeadlines(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	active := time.Now().Add(2 * time.Hour).Truncate(time.Second)
	if err := SaveSnoozedChats(map[string]time.Time{"active": active, "expired": time.Now().Add(-time.Hour)}); err != nil {
		t.Fatalf("SaveSnoozedChats: %v", err)
	}
	loaded := LoadSnoozedChats()
	if !loaded["active"].Equal(active) || len(loaded) != 1 {
		t.Fatalf("loaded snoozes = %#v", loaded)
	}
}

func TestConversationHeaderClearsTerminalRows(t *testing.T) {
	app := NewApp()
	name := "Current chat"
	app.SetChats([]Chat{{ID: "current", CachedDisplayName: &name}})
	app.SetSelectedChatID("current")
	model := NewModel(app, "client", "user")
	header := model.renderConversationHeader(40, "Direct chat")
	for _, line := range strings.Split(header, "\n") {
		if !strings.Contains(line, "\x1b[K") {
			t.Fatalf("header row does not clear stale terminal cells: %q", line)
		}
	}
}
