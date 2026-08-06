package main

import (
	"encoding/json"
	"testing"
)

func artifactEvent(id, eventType, createdAt, webURL string) Message {
	return Message{
		ID:              id,
		CreatedDateTime: createdAt,
		MessageType:     "systemEventMessage",
		WebURL:          webURL,
		EventDetail: &EventMessageDetail{
			ODataType: eventType,
		},
	}
}

func TestCollectConversationArtifactsPreservesDuplicateEvents(t *testing.T) {
	first := artifactEvent("recording-1", "#microsoft.graph.callRecordingEventMessageDetail", "2026-08-02T15:11:00Z", "https://teams.microsoft.com/l/message/one")
	first.EventDetail.CallRecordingURL = "https://contoso.sharepoint.com/recording-one"
	second := artifactEvent("recording-2", "#microsoft.graph.callRecordingEventMessageDetail", "2026-08-02T15:12:00Z", "https://teams.microsoft.com/l/message/two")
	second.EventDetail.CallRecordingURL = "https://contoso.sharepoint.com/recording-two"

	artifacts := collectConversationArtifacts([]Message{first, second}, "https://teams.microsoft.com/l/chat/fallback")
	if len(artifacts) != 2 {
		t.Fatalf("duplicate recording events were collapsed: %#v", artifacts)
	}
	if artifacts[0].URL != first.EventDetail.CallRecordingURL || artifacts[1].URL != second.EventDetail.CallRecordingURL {
		t.Fatalf("direct recording URLs were not retained: %#v", artifacts)
	}
}

func TestTranscriptArtifactUsesEventThenConversationFallback(t *testing.T) {
	withEventURL := artifactEvent("transcript-1", "#microsoft.graph.callTranscriptEventMessageDetail", "2026-08-02T15:11:00Z", "https://teams.microsoft.com/l/message/transcript")
	withoutEventURL := artifactEvent("transcript-2", "#microsoft.graph.callTranscriptEventMessageDetail", "2026-08-02T15:12:00Z", "")
	fallback := "https://teams.microsoft.com/l/chat/fallback"

	artifacts := collectConversationArtifacts([]Message{withEventURL, withoutEventURL}, fallback)
	if len(artifacts) != 2 {
		t.Fatalf("transcript artifacts = %#v", artifacts)
	}
	if artifacts[0].URL != withEventURL.WebURL || artifacts[0].DirectLink {
		t.Fatalf("transcript event URL fallback incorrect: %#v", artifacts[0])
	}
	if artifacts[1].URL != fallback {
		t.Fatalf("conversation fallback incorrect: %#v", artifacts[1])
	}
}

func TestEventDetailDecodesTranscriptIdentifiers(t *testing.T) {
	var detail EventMessageDetail
	if err := json.Unmarshal([]byte(`{"callId":"call-123","callTranscriptICalUid":"ical-456"}`), &detail); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if detail.CallID != "call-123" || detail.CallTranscriptICalUID != "ical-456" {
		t.Fatalf("transcript identifiers lost: %#v", detail)
	}
}
