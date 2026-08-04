package main

import (
	"reflect"
	"testing"
)

func messageIDs(messages []Message) []string {
	ids := make([]string, len(messages))
	for index, message := range messages {
		ids[index] = message.ID
	}
	return ids
}

func TestMessageOrderingUsesAbsoluteRFC3339Time(t *testing.T) {
	messages := []Message{
		{ID: "invalid", CreatedDateTime: "not-a-time"},
		{ID: "older-offset", CreatedDateTime: "2026-08-04T12:00:00+03:00"},
		{ID: "newer-fraction", CreatedDateTime: "2026-08-04T09:30:00.1000000Z"},
		{ID: "middle-fraction", CreatedDateTime: "2026-08-04T09:30:00.0900000Z"},
	}

	newest := append([]Message(nil), messages...)
	sortMessagesNewestFirst(newest)
	if got, want := messageIDs(newest), []string{"newer-fraction", "middle-fraction", "older-offset", "invalid"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("newest-first IDs = %v, want %v", got, want)
	}

	oldest := append([]Message(nil), messages...)
	sortMessagesOldestFirst(oldest)
	if got, want := messageIDs(oldest), []string{"older-offset", "middle-fraction", "newer-fraction", "invalid"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("oldest-first IDs = %v, want %v", got, want)
	}
}

func TestMessageOrderingBreaksEquivalentTimestampTiesByID(t *testing.T) {
	messages := []Message{
		{ID: "b", CreatedDateTime: "2026-08-04T12:00:00+03:00"},
		{ID: "a", CreatedDateTime: "2026-08-04T09:00:00Z"},
	}

	sortMessagesOldestFirst(messages)
	if got, want := messageIDs(messages), []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("oldest-first tied IDs = %v, want %v", got, want)
	}
	sortMessagesNewestFirst(messages)
	if got, want := messageIDs(messages), []string{"b", "a"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("newest-first tied IDs = %v, want %v", got, want)
	}
}
