package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestNormalizeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Jérémy", "Jeremy"},
		{"François", "Francois"},
		{"jeremy", "jeremy"},
		{"", ""},
		{"München", "Munchen"},
		{"Álvaro", "Alvaro"},
	}

	for _, test := range tests {
		result := normalizeString(test.input)
		if result != test.expected {
			t.Errorf("normalizeString(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}

func TestHighlightQuery(t *testing.T) {
	// Force a color profile so that lipgloss outputs ANSI sequences during headless tests.
	lipgloss.SetColorProfile(termenv.TrueColor)

	tests := []struct {
		text     string
		query    string
		contains string // what the highlighted text should contain
	}{
		{"Hello Jérémy", "jeremy", "Jérémy"},
		{"Hello Jeremy", "jeremy", "Jeremy"},
		{"François is here", "francois", "François"},
		{"No match here", "jeremy", "No match here"},
	}

	for _, test := range tests {
		result := highlightQuery(test.text, test.query)
		if test.text == test.contains {
			// If it's a "No match here", result should be identical to input
			if result != test.text {
				t.Errorf("highlightQuery(%q, %q) = %q; expected no match", test.text, test.query, result)
			}
		} else {
			// If it matches, the original substring (with accents) should be in the result,
			// wrapped in some ANSI escape codes.
			if !strings.Contains(result, test.contains) {
				t.Errorf("highlightQuery(%q, %q) = %q; expected to contain %q", test.text, test.query, result, test.contains)
			}
			// And it should not be identical to original text
			if result == test.text {
				t.Errorf("highlightQuery(%q, %q) = %q; expected highlighted output, got original text", test.text, test.query, result)
			}
		}
	}
}

func TestMessageMatches(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	msg := Message{
		ID: "1",
		From: &MessageFrom{
			User: &MessageUser{
				DisplayName: strPtr("Alice Smith"),
			},
		},
		Body: &MessageBody{
			Content: strPtr("<p>Hello world from Go</p>"),
		},
		Attachments: []MessageAttachment{
			{
				Name: strPtr("report.pdf"),
			},
		},
	}

	model := Model{}

	tests := []struct {
		query    string
		expected bool
	}{
		{"", true},
		{"hello", true},
		{"from Go", true},
		{"report", true},
		{"report.pdf", true},
		{"Alice", false}, // Should NOT match creator display name
		{"Smith", false}, // Should NOT match creator display name
		{"nonexistent", false},
	}

	for _, test := range tests {
		result := model.messageMatches(&msg, test.query)
		if result != test.expected {
			t.Errorf("messageMatches(query=%q) = %v; expected %v", test.query, result, test.expected)
		}
	}
}

func TestExtractAndProcessInlineImages(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	htmlContent := `<p>Check out this image: <img src="https://graph.microsoft.com/v1.0/chats/123/messages/456/hostedContents/abc/$value" alt="My screenshot" /> and another: <img src="https://graph.microsoft.com/v1.0/chats/123/messages/456/hostedContents/def/$value" /></p>`

	// Test ExtractInlineImages
	inlineAtts := ExtractInlineImages(htmlContent)
	if len(inlineAtts) != 2 {
		t.Fatalf("expected 2 inline images, got %d", len(inlineAtts))
	}

	if *inlineAtts[0].Name != "My screenshot.png" {
		t.Errorf("expected first name 'My screenshot.png', got %q", *inlineAtts[0].Name)
	}
	if *inlineAtts[0].ContentURL != "https://graph.microsoft.com/v1.0/chats/123/messages/456/hostedContents/abc/$value" {
		t.Errorf("expected first URL to match, got %q", *inlineAtts[0].ContentURL)
	}

	if *inlineAtts[1].Name != "inline-image-2.png" {
		t.Errorf("expected second name 'inline-image-2.png', got %q", *inlineAtts[1].Name)
	}
	if *inlineAtts[1].ContentURL != "https://graph.microsoft.com/v1.0/chats/123/messages/456/hostedContents/def/$value" {
		t.Errorf("expected second URL to match, got %q", *inlineAtts[1].ContentURL)
	}

	// Test ProcessInlineImages
	msg := Message{
		ID: "1",
		Body: &MessageBody{
			Content: strPtr(htmlContent),
		},
		Attachments: []MessageAttachment{
			{
				ID:         "existing-doc",
				Name:       strPtr("report.pdf"),
				ContentURL: strPtr("https://sharepoint.com/report.pdf"),
			},
		},
	}

	msg.ProcessInlineImages()

	if len(msg.Attachments) != 3 {
		t.Errorf("expected 3 attachments after processing inline images, got %d", len(msg.Attachments))
	}

	// Double processing check
	msg.ProcessInlineImages()
	if len(msg.Attachments) != 3 {
		t.Errorf("expected attachments count to remain 3 on double processing, got %d", len(msg.Attachments))
	}

	// Test HTMLToText inline image naming
	plainText := HTMLToText(htmlContent, msg.Attachments, nil)
	if !strings.Contains(plainText, "My screenshot.png") {
		t.Errorf("expected plainText to contain 'My screenshot.png', got %q", plainText)
	}
	if !strings.Contains(plainText, "inline-image-2.png") {
		t.Errorf("expected plainText to contain 'inline-image-2.png', got %q", plainText)
	}
}

func TestProcessInlineImagesResolvesTeamsHostedContentURLs(t *testing.T) {
	body := `<p>Screenshot <img src="../hostedContents/hosted-123/$value" alt="screen" /></p>`
	msg := Message{
		ID:     "message-456",
		ChatID: "19:chat@thread.v2",
		Body:   &MessageBody{Content: &body},
	}

	msg.ProcessInlineImages()
	if len(msg.Attachments) != 1 || msg.Attachments[0].ContentURL == nil {
		t.Fatalf("resolved attachments = %#v", msg.Attachments)
	}
	want := "https://graph.microsoft.com/v1.0/chats/19:chat@thread.v2/messages/message-456/hostedContents/hosted-123/$value"
	if got := *msg.Attachments[0].ContentURL; got != want {
		t.Fatalf("hosted content URL = %q, want %q", got, want)
	}
}

func TestProcessInlineImagesResolvesChannelReplyURLs(t *testing.T) {
	body := `<img src="../hostedContents/channel-image/$value" />`
	msg := Message{
		ID:        "reply-2",
		ReplyToID: "root-1",
		ChannelIdentity: &MessageChannelIdentity{
			TeamID:    "team-1",
			ChannelID: "channel-1",
		},
		Body: &MessageBody{Content: &body},
	}

	msg.ProcessInlineImages()
	want := "https://graph.microsoft.com/v1.0/teams/team-1/channels/channel-1/messages/root-1/replies/reply-2/hostedContents/channel-image/$value"
	if len(msg.Attachments) != 1 || msg.Attachments[0].ContentURL == nil || *msg.Attachments[0].ContentURL != want {
		t.Fatalf("channel hosted content URL = %#v, want %q", msg.Attachments, want)
	}
}

func TestProcessInlineImagesUpgradesRelativeCachedAttachment(t *testing.T) {
	body := `<img src="../hostedContents/hosted-123/$value" />`
	name := "inline-image-1.png"
	contentType := "image/png"
	relative := "../hostedContents/hosted-123/$value"
	msg := Message{
		ID:     "message-456",
		ChatID: "chat-1",
		Body:   &MessageBody{Content: &body},
		Attachments: []MessageAttachment{{
			ID:          "inline-img-1",
			Name:        &name,
			ContentType: &contentType,
			ContentURL:  &relative,
		}},
	}

	msg.ProcessInlineImages()
	if len(msg.Attachments) != 1 || msg.Attachments[0].ContentURL == nil || !strings.HasPrefix(*msg.Attachments[0].ContentURL, graphAPIBase) {
		t.Fatalf("cached inline attachment was not upgraded: %#v", msg.Attachments)
	}
}

func TestHTMLToTextMentions(t *testing.T) {
	// Force color profile for testing ANSI codes
	oldProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(oldProfile)

	// Helper functions for pointers
	intPtr := func(v int) *int { return &v }
	stringPtr := func(v string) *string { return &v }

	// Case 1: Mention without '@' prefix
	html1 := `Hello <at id="0">John Doe</at>!`
	res1 := HTMLToText(html1, nil, nil)
	plain1 := stripANSI(res1)
	expected1 := "Hello @John Doe!"
	if plain1 != expected1 {
		t.Errorf("expected %q, got %q", expected1, plain1)
	}
	// Check that ANSI styling was applied
	if !strings.Contains(res1, "\x1b[") {
		t.Errorf("expected res1 to contain ANSI escape codes, got %q", res1)
	}

	// Case 2: Mention that already starts with '@'
	html2 := `Hello <at id="1">@Jane Doe</at>!`
	res2 := HTMLToText(html2, nil, nil)
	plain2 := stripANSI(res2)
	expected2 := "Hello @Jane Doe!"
	if plain2 != expected2 {
		t.Errorf("expected %q, got %q", expected2, plain2)
	}

	// Case 3: Mention with non-breaking space
	html3 := `Hello <at id="2">John&nbsp;Doe</at>!`
	res3 := HTMLToText(html3, nil, nil)
	plain3 := stripANSI(res3)
	expected3 := "Hello @John\u00a0Doe!" // \u00a0 is nbsp
	if plain3 != expected3 {
		t.Errorf("expected %q, got %q", expected3, plain3)
	}

	// Case 4: Mention split into two tags with same ID
	html4 := `Hello <at id="0">John</at> <at id="0">Doe</at>!`
	res4 := HTMLToText(html4, nil, nil)
	plain4 := stripANSI(res4)
	expected4 := "Hello @John Doe!"
	if plain4 != expected4 {
		t.Errorf("expected %q, got %q", expected4, plain4)
	}

	// Case 5: Mention split into two tags with different IDs but same user ID
	html5 := `Hello <at id="0">John</at> <at id="1">Doe</at>!`
	mentions5 := []MessageMention{
		{
			ID:          intPtr(0),
			MentionText: stringPtr("John"),
			Mentioned: &MentionedIdentitySet{
				User: &MessageUser{
					ID:          stringPtr("user-123"),
					DisplayName: stringPtr("John Doe"),
				},
			},
		},
		{
			ID:          intPtr(1),
			MentionText: stringPtr("Doe"),
			Mentioned: &MentionedIdentitySet{
				User: &MessageUser{
					ID:          stringPtr("user-123"),
					DisplayName: stringPtr("John Doe"),
				},
			},
		},
	}
	res5 := HTMLToText(html5, nil, mentions5)
	plain5 := stripANSI(res5)
	expected5 := "Hello @John Doe!"
	if plain5 != expected5 {
		t.Errorf("expected %q, got %q", expected5, plain5)
	}
}

func TestComputeDisplayName(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name     string
		chat     Chat
		expected string
	}{
		{
			name: "One-on-one chat",
			chat: Chat{
				ChatType: "oneOnOne",
				Members: []ChatMember{
					{DisplayName: strPtr("Alice Smith")},
				},
			},
			expected: "Alice Smith",
		},
		{
			name: "Group chat with topic",
			chat: Chat{
				ChatType: "group",
				Topic:    strPtr("Project Alpha"),
				Members: []ChatMember{
					{DisplayName: strPtr("Alice Smith")},
					{DisplayName: strPtr("Bob Jones")},
				},
			},
			expected: "Project Alpha",
		},
		{
			name: "Group chat with 2 members",
			chat: Chat{
				ChatType: "group",
				Members: []ChatMember{
					{DisplayName: strPtr("Alice Smith")},
					{DisplayName: strPtr("Bob Jones")},
				},
			},
			expected: "Alice S, Bob J",
		},
		{
			name: "Group chat with 3 members",
			chat: Chat{
				ChatType: "group",
				Members: []ChatMember{
					{DisplayName: strPtr("Alice Smith")},
					{DisplayName: strPtr("Bob Jones")},
					{DisplayName: strPtr("Charlie Brown")},
				},
			},
			expected: "Alice S, Bob J, Charlie B",
		},
		{
			name: "Group chat with 4 members",
			chat: Chat{
				ChatType: "group",
				Members: []ChatMember{
					{DisplayName: strPtr("Alice Smith")},
					{DisplayName: strPtr("Bob Jones")},
					{DisplayName: strPtr("Charlie Brown")},
					{DisplayName: strPtr("David Miller")},
				},
			},
			expected: "Alice S, Bob J, Charlie B ...",
		},
		{
			name: "Group chat with some nil DisplayNames",
			chat: Chat{
				ChatType: "group",
				Members: []ChatMember{
					{DisplayName: strPtr("Alice Smith")},
					{DisplayName: nil},
					{DisplayName: strPtr("Bob Jones")},
				},
			},
			expected: "Alice S, Bob J",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeDisplayName(&tt.chat)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFilterCurrentMemberPrefersAuthenticatedUserID(t *testing.T) {
	currentName := "Alex Smith"
	selfID := "current-user"
	otherID := "other-user"
	members := []ChatMember{
		{UserID: &selfID, DisplayName: &currentName},
		{UserID: &otherID, DisplayName: &currentName},
	}

	filtered := filterCurrentMember(members, selfID, &currentName)
	if len(filtered) != 1 || filtered[0].UserID == nil || *filtered[0].UserID != otherID {
		t.Fatalf("filtered members = %#v, want only the non-current member", filtered)
	}
	if len(members) != 2 {
		t.Fatalf("filterCurrentMember mutated its input: %#v", members)
	}
}

func TestFilterCurrentMemberFallsBackToDisplayName(t *testing.T) {
	currentName := "MDF"
	otherName := "Ada Lovelace"
	members := []ChatMember{
		{DisplayName: &currentName},
		{DisplayName: &otherName},
	}

	filtered := filterCurrentMember(members, "current-user", &currentName)
	if len(filtered) != 1 || filtered[0].DisplayName == nil || *filtered[0].DisplayName != otherName {
		t.Fatalf("filtered members = %#v, want only Ada Lovelace", filtered)
	}
}

func TestFilterMessageAttachments(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	cardType := "application/vnd.microsoft.card.thumbnail"
	referenceType := "reference"
	imageType := "image/png"
	messageRefType := "messageReference"

	msg := Message{
		Attachments: []MessageAttachment{
			{
				ID:          "1",
				ContentType: &cardType,
				ContentURL:  strPtr("https://www.youtube.com/watch?v=123"),
			},
			{
				ID:          "2",
				ContentType: &referenceType,
				ContentURL:  strPtr("https://some-tenant.sharepoint.com/personal/file.docx"),
			},
			{
				ID:          "3",
				ContentType: &referenceType,
				ContentURL:  strPtr("https://www.youtube.com/watch?v=456"),
			},
			{
				ID:          "4",
				ContentType: &imageType,
				ContentURL:  strPtr("https://graph.microsoft.com/v1.0/chats/inline-img"),
			},
			{
				ID:          "5",
				ContentType: &messageRefType,
			},
		},
	}

	FilterMessageAttachments(&msg)

	if len(msg.Attachments) != 3 {
		t.Fatalf("expected 3 attachments after filtering, got %d", len(msg.Attachments))
	}

	expectedIDs := map[string]bool{"2": true, "4": true, "5": true}
	for _, att := range msg.Attachments {
		if !expectedIDs[att.ID] {
			t.Errorf("unexpected attachment ID remaining: %s", att.ID)
		}
	}
}

func TestGetAttachmentSavedName(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name        string
		att         MessageAttachment
		defaultName string
		expected    string
	}{
		{
			name: "Pasted image with ID and extension",
			att: MessageAttachment{
				ID:   "inline-img-1",
				Name: strPtr("image.png"),
			},
			defaultName: "image",
			expected:    "image_inline-img-1.png",
		},
		{
			name: "Attachment with custom name and numeric ID",
			att: MessageAttachment{
				ID:   "123",
				Name: strPtr("screenshot.jpg"),
			},
			defaultName: "image",
			expected:    "screenshot_123.jpg",
		},
		{
			name: "Attachment with nil name using default name and ID",
			att: MessageAttachment{
				ID: "att-456",
			},
			defaultName: "attachment",
			expected:    "attachment_att-456",
		},
		{
			name: "Stem already has ID suffix",
			att: MessageAttachment{
				ID:   "123",
				Name: strPtr("image_123.png"),
			},
			defaultName: "image",
			expected:    "image_123.png",
		},
		{
			name: "No ID but has ContentURL",
			att: MessageAttachment{
				Name:       strPtr("file.txt"),
				ContentURL: strPtr("https://example.com/file.txt"),
			},
			defaultName: "attachment",
			// URL sha256 prefix will be appended
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getAttachmentSavedName(tt.att, tt.defaultName)
			if tt.expected != "" && got != tt.expected {
				t.Errorf("getAttachmentSavedName() = %q, expected %q", got, tt.expected)
			}
			if tt.expected == "" && got == "file.txt" {
				t.Errorf("getAttachmentSavedName() expected unique hash appended, got %q", got)
			}
		})
	}
}
