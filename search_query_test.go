package main

import (
	"testing"
	"time"
)

func TestParseSearchQuerySupportsQuotesFieldsAndNegation(t *testing.T) {
	query := parseSearchQuery(`"quarterly plan" from:"Alice Smith" is:unread -has:file`)
	if len(query.Terms) != 4 {
		t.Fatalf("terms = %#v", query.Terms)
	}
	if query.Terms[0].Field != "" || query.Terms[0].Value != "quarterly plan" {
		t.Fatalf("quoted phrase parsed incorrectly: %#v", query.Terms[0])
	}
	if query.Terms[1].Field != "from" || query.Terms[1].Value != "Alice Smith" {
		t.Fatalf("quoted field parsed incorrectly: %#v", query.Terms[1])
	}
	if !query.Terms[3].Negated || query.Terms[3].Field != "has" {
		t.Fatalf("negated field parsed incorrectly: %#v", query.Terms[3])
	}
}

func TestFuzzyTextScoreRequiresOrderedCharacters(t *testing.T) {
	if _, matched := fuzzyTextScore("Quarterly planning", "qtrly"); !matched {
		t.Fatal("ordered fuzzy abbreviation did not match")
	}
	if _, matched := fuzzyTextScore("Quarterly planning", "ylrtq"); matched {
		t.Fatal("out-of-order abbreviation matched")
	}
	if _, matched := fuzzyTextScore("Alice", "ac"); matched {
		t.Fatal("two-character non-substring query should not fuzzy-match")
	}
}

func TestStructuredMessageQuery(t *testing.T) {
	target := searchTarget{
		Text:         []string{"Quarterly roadmap approved"},
		Sender:       []string{"Alice Smith"},
		Conversation: []string{"Product planning"},
		Kind:         "message",
		CreatedAt:    time.Date(2026, 8, 5, 12, 0, 0, 0, time.Local),
		Unread:       true,
		Favorite:     true,
		HasFile:      true,
	}
	query := parseSearchQuery(`qtrly from:alice in:product is:unread is:favorite has:file after:2026-08-01 -type:event`)
	if _, matched := query.Match(target); !matched {
		t.Fatal("structured fuzzy query did not match target")
	}
	if _, matched := parseSearchQuery(`before:2026-08-01`).Match(target); matched {
		t.Fatal("date exclusion did not apply")
	}
}

func TestGlobalSearchUsesHiddenChatsAndAlreadyLoadedMessages(t *testing.T) {
	app := NewApp()
	visibleName := "Visible"
	hiddenName := "Planning archive"
	app.SetChats([]Chat{{ID: "visible", CachedDisplayName: &visibleName}})
	model := NewModel(app, "client", "user")
	hidden := Chat{ID: "hidden", CachedDisplayName: &hiddenName}
	model.chatCache[hidden.ID] = hidden
	model.stableChatOrder = []string{"visible", "hidden"}
	body := "The quarterly roadmap is ready"
	app.CachedMessages[hidden.ID] = []Message{{
		ID:              "message-1",
		CreatedDateTime: "2026-08-05T12:00:00Z",
		Body:            &MessageBody{Content: &body},
	}}

	app.UserSearchQuery = "qtrly"
	model.updateUserSearchLocalResults()
	if len(app.UserSearchMessageResults) != 1 || app.UserSearchMessageResults[0].ChatID != hidden.ID {
		t.Fatalf("hidden loaded message missing from search: %#v", app.UserSearchMessageResults)
	}
	if len(app.CachedMessages[hidden.ID]) != 1 {
		t.Fatal("search mutated the existing message cache")
	}
}
