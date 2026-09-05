package main

import (
	"testing"
	"time"
)

func TestChatReadStatePayload(t *testing.T) {
	payload := chatReadStatePayload("user-1", "tenant-1")
	user, ok := payload["user"].(map[string]string)
	if !ok {
		t.Fatalf("expected user identity payload, got %#v", payload)
	}
	if user["id"] != "user-1" || user["tenantId"] != "tenant-1" {
		t.Fatalf("unexpected user identity: %#v", user)
	}
	if _, present := payload["lastMessageReadDateTime"]; present {
		t.Fatal("default unread action must omit lastMessageReadDateTime")
	}
}

func TestOwnMessagePreservesUnreadState(t *testing.T) {
	user := "Me"
	other := "Other"
	old := filterTestMessage("old", other)
	newest := filterTestMessage("new", user)
	newest.CreatedDateTime = "2026-08-02T10:05:00Z"
	chat := Chat{ID: "chat-1", LastMessagePreview: old}
	app := NewApp()
	app.CurrentUserName = &user
	app.Chats = []Chat{chat}
	app.SelectedIndex = 0
	model := NewModel(app, "client", "user")
	model.latestChats = []Chat{chat}
	model.stableChatOrder = []string{chat.ID}
	model.lastMsgID[chat.ID] = old.ID
	model.lastMsgTime[chat.ID], _ = time.Parse(time.RFC3339, old.CreatedDateTime)
	model.lastReadMsgID[chat.ID] = "earlier"

	chat.LastMessagePreview = newest
	model, _ = model.updateInternal(MsgChatsLoaded{Chats: []Chat{chat}})
	if !model.isUnread(chat) {
		t.Fatal("outgoing message changed an unread chat to read")
	}
	if got := model.lastReadMsgID[chat.ID]; got == newest.ID {
		t.Fatalf("outgoing message advanced read marker to %q", got)
	}
}

func TestOwnMessagePreservesReadState(t *testing.T) {
	user := "Me"
	old := filterTestMessage("old", user)
	newest := filterTestMessage("new", user)
	newest.CreatedDateTime = "2026-08-02T10:05:00Z"
	chat := Chat{ID: "chat-1", LastMessagePreview: old}
	app := NewApp()
	app.CurrentUserName = &user
	app.Chats = []Chat{chat}
	model := NewModel(app, "client", "user")
	model.lastMsgID[chat.ID] = old.ID
	model.lastMsgTime[chat.ID], _ = time.Parse(time.RFC3339, old.CreatedDateTime)
	model.lastReadMsgID[chat.ID] = old.ID

	chat.LastMessagePreview = newest
	model, _ = model.updateInternal(MsgChatsLoaded{Chats: []Chat{chat}})
	if got := model.lastReadMsgID[chat.ID]; got != newest.ID {
		t.Fatalf("read chat did not remain read: marker=%q, want %q", got, newest.ID)
	}
}

func TestMarkReadRespectsNavigationPolicy(t *testing.T) {
	chat := Chat{ID: "chat-1"}
	app := NewApp()
	app.Chats = []Chat{chat}
	app.SelectedIndex = 0
	app.MarkReadOnOpen = false
	model := NewModel(app, "client", "user")
	model.focused = true
	model.lastMsgID[chat.ID] = "message-1"

	model = model.markRead()
	if _, present := model.lastReadMsgID[chat.ID]; present {
		t.Fatal("navigation marked a chat read while mark_read_on_open was false")
	}

	app.MarkReadOnOpen = true
	model.manuallyUnread[chat.ID] = true
	model = model.markRead()
	if _, present := model.lastReadMsgID[chat.ID]; present {
		t.Fatal("auto-read overrode an explicitly unread chat before navigation")
	}
}

func TestChangedServerViewpointReconcilesExternalRead(t *testing.T) {
	message := filterTestMessage("latest", "Other")
	message.CreatedDateTime = "2026-08-02T10:00:00Z"
	chat := Chat{ID: "chat-1", LastMessagePreview: message,
		Viewpoint: &ChatViewpoint{LastMessageReadDateTime: "2026-08-02T09:00:00Z"}}
	model := NewModel(NewApp(), "client", "user")
	model.latestChats = []Chat{chat}
	model.lastMsgID[chat.ID] = message.ID
	model.lastMsgTime[chat.ID], _ = time.Parse(time.RFC3339, message.CreatedDateTime)
	model.reconcileServerReadState(chat)
	if !model.isUnread(chat) {
		t.Fatal("initial server viewpoint did not establish unread state")
	}

	chat.Viewpoint.LastMessageReadDateTime = "2026-08-02T10:01:00Z"
	model.reconcileServerReadState(chat)
	if model.isUnread(chat) || model.lastReadMsgID[chat.ID] != message.ID {
		t.Fatal("newer Teams viewpoint did not reconcile chat to read")
	}
}

func TestUnchangedServerViewpointDoesNotRollBackLocalRead(t *testing.T) {
	message := filterTestMessage("latest", "Other")
	message.CreatedDateTime = "2026-08-02T10:00:00Z"
	chat := Chat{ID: "chat-1", LastMessagePreview: message,
		Viewpoint: &ChatViewpoint{LastMessageReadDateTime: "2026-08-02T09:00:00Z"}}
	model := NewModel(NewApp(), "client", "user")
	model.lastMsgID[chat.ID] = message.ID
	model.lastMsgTime[chat.ID], _ = time.Parse(time.RFC3339, message.CreatedDateTime)
	model.reconcileServerReadState(chat)
	model.lastReadMsgID[chat.ID] = message.ID

	model.reconcileServerReadState(chat)
	if model.isUnread(chat) {
		t.Fatal("unchanged, eventually consistent viewpoint rolled back local read")
	}
}

func TestChangedServerViewpointReconcilesExternalUnread(t *testing.T) {
	message := filterTestMessage("latest", "Other")
	message.CreatedDateTime = "2026-08-02T10:00:00Z"
	chat := Chat{ID: "chat-1", LastMessagePreview: message,
		Viewpoint: &ChatViewpoint{LastMessageReadDateTime: "2026-08-02T10:01:00Z"}}
	model := NewModel(NewApp(), "client", "user")
	model.lastMsgID[chat.ID] = message.ID
	model.lastMsgTime[chat.ID], _ = time.Parse(time.RFC3339, message.CreatedDateTime)
	model.reconcileServerReadState(chat)

	chat.Viewpoint.LastMessageReadDateTime = "2026-08-02T09:00:00Z"
	model.reconcileServerReadState(chat)
	if !model.isUnread(chat) {
		t.Fatal("older changed Teams viewpoint did not reconcile chat to unread")
	}
}

func TestServerChatReadStateReportsUnavailableAndReadState(t *testing.T) {
	if got := serverChatReadState(Chat{}); got != "unavailable" {
		t.Fatalf("missing viewpoint state = %q", got)
	}
	message := filterTestMessage("latest", "Other")
	message.CreatedDateTime = "2026-08-02T10:00:00Z"
	chat := Chat{LastMessagePreview: message,
		Viewpoint: &ChatViewpoint{LastMessageReadDateTime: "2026-08-02T10:01:00Z"}}
	if got := serverChatReadState(chat); got != "read" {
		t.Fatalf("read viewpoint state = %q", got)
	}
	chat.Viewpoint.LastMessageReadDateTime = "2026-08-02T09:00:00Z"
	if got := serverChatReadState(chat); got != "unread" {
		t.Fatalf("unread viewpoint state = %q", got)
	}
	chat.Viewpoint.LastMessageReadDateTime = ""
	if got := serverChatReadState(chat); got != "unread" {
		t.Fatalf("cleared mark-unread viewpoint state = %q", got)
	}
}
