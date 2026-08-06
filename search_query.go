package main

import (
	"sort"
	"strings"
	"time"
	"unicode"
)

type searchQueryTerm struct {
	Field   string
	Value   string
	Negated bool
}

type searchQuery struct {
	Terms []searchQueryTerm
}

type searchTarget struct {
	Text         []string
	Sender       []string
	Conversation []string
	Kind         string
	CreatedAt    time.Time
	Unread       bool
	Favorite     bool
	HasFile      bool
	HasImage     bool
	HasLink      bool
}

type MessageSearchResult struct {
	ChatID   string
	ChatName string
	Message  Message
	Score    int
}

var queryFields = map[string]bool{
	"from":   true,
	"in":     true,
	"is":     true,
	"type":   true,
	"has":    true,
	"after":  true,
	"before": true,
}

func tokenizeSearchQuery(input string) []string {
	var tokens []string
	var current strings.Builder
	quoted := false
	escaped := false
	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}
	for _, r := range input {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == '"' {
			quoted = !quoted
			continue
		}
		if unicode.IsSpace(r) && !quoted {
			flush()
			continue
		}
		current.WriteRune(r)
	}
	if escaped {
		current.WriteRune('\\')
	}
	flush()
	return tokens
}

func parseSearchQuery(input string) searchQuery {
	query := searchQuery{}
	for _, raw := range tokenizeSearchQuery(input) {
		term := searchQueryTerm{}
		if strings.HasPrefix(raw, "-") && len(raw) > 1 {
			term.Negated = true
			raw = strings.TrimPrefix(raw, "-")
		}
		if field, value, found := strings.Cut(raw, ":"); found && queryFields[strings.ToLower(field)] && value != "" {
			term.Field = strings.ToLower(field)
			term.Value = value
		} else {
			term.Value = raw
		}
		term.Value = strings.TrimSpace(term.Value)
		if term.Value != "" {
			query.Terms = append(query.Terms, term)
		}
	}
	return query
}

func normalizedSearchText(text string) string {
	return normalizeString(strings.ToLower(strings.TrimSpace(text)))
}

func fuzzyTextScore(text, needle string) (int, bool) {
	haystack := []rune(normalizedSearchText(text))
	wanted := []rune(normalizedSearchText(needle))
	if len(wanted) == 0 {
		return 0, true
	}
	if len(haystack) == 0 {
		return 0, false
	}
	if index := strings.Index(string(haystack), string(wanted)); index >= 0 {
		return 1000 - index*2 - max(0, len(haystack)-len(wanted)), true
	}
	if len(wanted) < 3 {
		return 0, false
	}

	wantedIndex := 0
	start := -1
	last := -1
	gaps := 0
	consecutive := 0
	for index, r := range haystack {
		if r != wanted[wantedIndex] {
			continue
		}
		if start < 0 {
			start = index
		}
		if last >= 0 {
			if index == last+1 {
				consecutive++
			} else {
				gaps += index - last - 1
			}
		}
		last = index
		wantedIndex++
		if wantedIndex == len(wanted) {
			score := 300 + consecutive*12 - gaps*4 - start*2
			if score < 1 {
				score = 1
			}
			return score, true
		}
	}
	return 0, false
}

func fuzzyFieldsScore(fields []string, value string) (int, bool) {
	best := 0
	matched := false
	for _, field := range fields {
		if score, ok := fuzzyTextScore(field, value); ok {
			matched = true
			if score > best {
				best = score
			}
		}
	}
	return best, matched
}

func normalizedSearchKind(kind string) string {
	switch strings.ToLower(kind) {
	case "oneonone", "direct", "1:1":
		return "direct"
	case "group":
		return "group"
	case "meeting":
		return "meeting"
	case "event", "systemeventmessage":
		return "event"
	default:
		return strings.ToLower(kind)
	}
}

func (term searchQueryTerm) match(target searchTarget) (int, bool) {
	value := normalizedSearchText(term.Value)
	switch term.Field {
	case "from":
		return fuzzyFieldsScore(target.Sender, value)
	case "in":
		return fuzzyFieldsScore(target.Conversation, value)
	case "is":
		switch value {
		case "unread":
			return 100, target.Unread
		case "read":
			return 100, !target.Unread
		case "favorite", "favourite":
			return 100, target.Favorite
		}
		return 0, false
	case "type":
		return 100, normalizedSearchKind(target.Kind) == normalizedSearchKind(value)
	case "has":
		switch value {
		case "file", "attachment":
			return 100, target.HasFile
		case "image":
			return 100, target.HasImage
		case "link":
			return 100, target.HasLink
		}
		return 0, false
	case "after", "before":
		day, err := time.ParseInLocation("2006-01-02", value, time.Local)
		if err != nil || target.CreatedAt.IsZero() {
			return 0, false
		}
		if term.Field == "after" {
			return 100, target.CreatedAt.After(day)
		}
		return 100, target.CreatedAt.Before(day)
	default:
		return fuzzyFieldsScore(target.Text, value)
	}
}

func (query searchQuery) Match(target searchTarget) (int, bool) {
	total := 0
	for _, term := range query.Terms {
		score, matched := term.match(target)
		if term.Negated {
			if matched {
				return 0, false
			}
			continue
		}
		if !matched {
			return 0, false
		}
		total += score
	}
	return total, true
}

func (query searchQuery) HighlightTerms() []string {
	var terms []string
	for _, term := range query.Terms {
		if !term.Negated && (term.Field == "" || term.Field == "from" || term.Field == "in") {
			terms = append(terms, term.Value)
		}
	}
	return terms
}

func searchTime(value string) time.Time {
	when, _ := time.Parse(time.RFC3339Nano, value)
	return when.Local()
}

func attachmentSearchState(attachments []MessageAttachment) (hasFile, hasImage, hasLink bool) {
	for _, attachment := range attachments {
		contentType := ""
		if attachment.ContentType != nil {
			contentType = strings.ToLower(*attachment.ContentType)
		}
		if !strings.EqualFold(contentType, "messageReference") {
			hasFile = true
		}
		if strings.HasPrefix(contentType, "image/") {
			hasImage = true
		}
		if attachment.ContentURL != nil && strings.TrimSpace(*attachment.ContentURL) != "" {
			hasLink = true
		}
	}
	return
}

func messageSearchTarget(message *Message, chat Chat, unread, favorite bool) searchTarget {
	name := ""
	if chat.CachedDisplayName != nil {
		name = *chat.CachedDisplayName
	}
	hasFile, hasImage, hasLink := attachmentSearchState(message.Attachments)
	body := message.GetPlainText()
	text := []string{body, message.Subject, message.Summary}
	for _, attachment := range message.Attachments {
		if attachment.Name != nil {
			text = append(text, *attachment.Name)
		}
	}
	if strings.Contains(strings.ToLower(body), "http://") || strings.Contains(strings.ToLower(body), "https://") {
		hasLink = true
	}
	kind := "message"
	if message.IsSystemEvent() {
		kind = "event"
	}
	return searchTarget{
		Text:         text,
		Sender:       []string{message.SenderName()},
		Conversation: []string{name, chat.ID},
		Kind:         kind,
		CreatedAt:    searchTime(message.CreatedDateTime),
		Unread:       unread,
		Favorite:     favorite,
		HasFile:      hasFile,
		HasImage:     hasImage,
		HasLink:      hasLink,
	}
}

func chatSearchTarget(chat Chat, unread, favorite bool) searchTarget {
	var text []string
	var senders []string
	var conversation []string
	if chat.CachedDisplayName != nil {
		text = append(text, *chat.CachedDisplayName)
		conversation = append(conversation, *chat.CachedDisplayName)
	}
	if chat.Topic != nil {
		text = append(text, *chat.Topic)
		conversation = append(conversation, *chat.Topic)
	}
	conversation = append(conversation, chat.ID)
	for _, member := range chat.Members {
		if member.DisplayName != nil {
			text = append(text, *member.DisplayName)
		}
		if member.Email != nil {
			text = append(text, *member.Email)
		}
	}
	createdAt := time.Time{}
	hasFile, hasImage, hasLink := false, false, false
	if chat.LastMessagePreview != nil {
		preview := chat.LastMessagePreview
		text = append(text, preview.GetPlainText(), preview.Subject, preview.Summary)
		senders = append(senders, preview.SenderName())
		createdAt = searchTime(preview.CreatedDateTime)
		hasFile, hasImage, hasLink = attachmentSearchState(preview.Attachments)
		body := strings.ToLower(preview.GetPlainText())
		if strings.Contains(body, "http://") || strings.Contains(body, "https://") {
			hasLink = true
		}
	} else if chat.LastUpdated != nil {
		createdAt = searchTime(*chat.LastUpdated)
	}
	return searchTarget{
		Text:         text,
		Sender:       senders,
		Conversation: conversation,
		Kind:         chat.ChatType,
		CreatedAt:    createdAt,
		Unread:       unread,
		Favorite:     favorite,
		HasFile:      hasFile,
		HasImage:     hasImage,
		HasLink:      hasLink,
	}
}

func sortMessageSearchResults(results []MessageSearchResult) {
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return searchTime(results[i].Message.CreatedDateTime).After(searchTime(results[j].Message.CreatedDateTime))
	})
}
