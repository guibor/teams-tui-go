package main

import (
	"strings"
	"testing"
	"time"
)

func conversationViewMessage(id, sender, body, createdAt string, reply bool) Message {
	senderID := "id-" + sender
	return Message{
		ID:              id,
		CreatedDateTime: createdAt,
		From: &MessageFrom{User: &MessageUser{
			ID:          &senderID,
			DisplayName: &sender,
		}},
		Body:    &MessageBody{Content: &body},
		IsReply: reply,
	}
}

func TestConversationViewShowsDaySeparators(t *testing.T) {
	app := NewApp()
	app.Messages = []Message{
		conversationViewMessage("new", "Alice", "New", "2026-08-05T12:00:00Z", false),
		conversationViewMessage("old", "Alice", "Old", "2026-08-04T12:00:00Z", false),
	}
	model := NewModel(app, "client", "user")
	rendered := stripANSI(model.renderMessages(80, 30))
	for _, raw := range []string{"2026-08-04T12:00:00Z", "2026-08-05T12:00:00Z"} {
		when, _ := time.Parse(time.RFC3339, raw)
		label := when.Local().Format("Mon, Jan 02, 2006")
		if !strings.Contains(rendered, label) {
			t.Fatalf("missing day separator %q in:\n%s", label, rendered)
		}
	}
}

func TestConversationViewDoesNotGroupReplyHeaders(t *testing.T) {
	app := NewApp()
	app.Messages = []Message{
		conversationViewMessage("reply-2", "Alice", "Second reply", "2026-08-05T12:02:00Z", true),
		conversationViewMessage("reply-1", "Alice", "First reply", "2026-08-05T12:01:00Z", true),
	}
	model := NewModel(app, "client", "user")
	rendered := stripANSI(model.renderMessages(80, 30))
	if count := strings.Count(rendered, "Alice"); count != 2 {
		t.Fatalf("reply sender headers = %d, want 2:\n%s", count, rendered)
	}
}

func TestConversationViewKeepsDuplicateSystemEvents(t *testing.T) {
	first := artifactEvent("event-1", "#microsoft.graph.callRecordingEventMessageDetail", "2026-08-05T12:01:00Z", "")
	second := artifactEvent("event-2", "#microsoft.graph.callRecordingEventMessageDetail", "2026-08-05T12:02:00Z", "")
	app := NewApp()
	app.Messages = []Message{second, first}
	model := NewModel(app, "client", "user")
	rendered := stripANSI(model.renderMessages(80, 30))
	if count := strings.Count(rendered, "Call recording available"); count != 2 {
		t.Fatalf("system event count = %d, want 2:\n%s", count, rendered)
	}
}

func TestConversationMetadataSummarizesStateAndResources(t *testing.T) {
	name := "Planning"
	members := []ChatMember{{}, {}}
	chat := Chat{ID: "chat", ChatType: "group", CachedDisplayName: &name, Members: members}
	event := artifactEvent("event", "#microsoft.graph.callTranscriptEventMessageDetail", "2026-08-05T12:02:00Z", "https://teams.microsoft.com/l/message/transcript")
	app := NewApp()
	app.SetChats([]Chat{chat})
	app.SetSelectedChatID(chat.ID)
	app.SetMessages(chat.ID, []Message{event}, "")
	model := NewModel(app, "client", "user")
	model.chatCache[chat.ID] = chat
	model.favourites[chat.ID] = true

	detail := model.conversationMetadataDetail()
	for _, expected := range []string{"Group chat", "2 members", "1 loaded message", "Favorite", "1 transcript"} {
		if !strings.Contains(detail, expected) {
			t.Fatalf("metadata missing %q: %q", expected, detail)
		}
	}
}
