package main

import "testing"

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
