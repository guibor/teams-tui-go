package main

import (
	"strings"
	"testing"
)

func TestBidiVisualLineMixedEnglishAndHebrew(t *testing.T) {
	visual, rtl := bidiVisualLine("Hello שלום world")
	if rtl {
		t.Fatal("English-first mixed text should keep LTR paragraph alignment")
	}
	if visual != "Hello םולש world" {
		t.Fatalf("unexpected mixed-direction output: %q", visual)
	}
}

func TestBidiVisualLineHebrewParagraph(t *testing.T) {
	visual, rtl := bidiVisualLine("שלום Hello עולם")
	if !rtl {
		t.Fatal("Hebrew-first text should request RTL paragraph alignment")
	}
	if visual != "םלוע Hello םולש" {
		t.Fatalf("unexpected RTL paragraph output: %q", visual)
	}
}

func TestBidiVisualLineNumbersAndBrackets(t *testing.T) {
	tests := []struct {
		logical string
		visual  string
	}{
		{logical: "שלום 123", visual: "123 םולש"},
		{logical: "שלום (123)", visual: "(123) םולש"},
		{logical: "English (שלום) text", visual: "English (םולש) text"},
	}
	for _, test := range tests {
		actual, _ := bidiVisualLine(test.logical)
		if actual != test.visual {
			t.Errorf("bidiVisualLine(%q) = %q, want %q", test.logical, actual, test.visual)
		}
	}
}

func TestBidiVisualLinePreservesGraphemeClusters(t *testing.T) {
	visual, _ := bidiVisualLine("שָׁלוֹם")
	if visual != "םוֹלשָׁ" {
		t.Fatalf("Hebrew marks detached from their base letters: %q", visual)
	}
}

func TestBidiVisualLinePreservesANSIStyles(t *testing.T) {
	visual, _ := bidiVisualLine("\x1b[1mשלום\x1b[0m world")
	tokens := parseANSIStyledRunes(visual)
	if plainStyledRunes(tokens) != "world םולש" {
		t.Fatalf("unexpected styled visual text: %q", plainStyledRunes(tokens))
	}
	for _, token := range tokens {
		if strings.ContainsRune("םולש", token.r) && token.style.sgr == "" {
			t.Fatalf("Hebrew rune %q lost its ANSI style in %q", token.r, visual)
		}
	}
}

func TestBidiVisualLinePreservesOSC8Hyperlinks(t *testing.T) {
	logical := "Open \x1b]8;;https://example.test\x1b\\שלום\x1b]8;;\x1b\\ now"
	visual, _ := bidiVisualLine(logical)
	tokens := parseANSIStyledRunes(visual)
	if plainStyledRunes(tokens) != "Open םולש now" {
		t.Fatalf("unexpected linked visual text: %q", plainStyledRunes(tokens))
	}
	for _, token := range tokens {
		if strings.ContainsRune("םולש", token.r) && token.style.link == "" {
			t.Fatalf("Hebrew rune %q lost its OSC 8 link in %q", token.r, visual)
		}
	}
}

func TestBidiVisualLineLeavesLTRANSIUntouched(t *testing.T) {
	logical := "\x1b[1mplain English\x1b[0m"
	visual, rtl := bidiVisualLine(logical)
	if rtl || visual != logical {
		t.Fatalf("LTR-only content changed: %q", visual)
	}
}

func TestRenderMessagesRightAlignsHebrewFirstLines(t *testing.T) {
	logical := "שלום עולם"
	sender := "Colleague"
	app := NewApp()
	app.Messages = []Message{{
		ID:              "message-1",
		CreatedDateTime: "2026-08-02T12:00:00Z",
		From:            &MessageFrom{User: &MessageUser{DisplayName: &sender}},
		PlainTextCached: &logical,
	}}
	model := NewModel(app, "client", "user")

	output := model.renderMessages(40, 10)
	bodyLine := findLineContaining(output, "םלוע םולש")
	if bodyLine == "" {
		t.Fatalf("rendered output did not contain visually ordered Hebrew: %q", output)
	}
	if !strings.HasPrefix(bodyLine, " ") {
		t.Fatalf("Hebrew-first body was not right-aligned: %q", bodyLine)
	}
}

func TestRenderMessagesKeepsEnglishFirstMixedLinesLeftAligned(t *testing.T) {
	logical := "Hello שלום world"
	sender := "Colleague"
	app := NewApp()
	app.Messages = []Message{{
		ID:              "message-1",
		CreatedDateTime: "2026-08-02T12:00:00Z",
		From:            &MessageFrom{User: &MessageUser{DisplayName: &sender}},
		PlainTextCached: &logical,
	}}
	model := NewModel(app, "client", "user")

	output := model.renderMessages(40, 10)
	bodyLine := findLineContaining(output, "Hello םולש world")
	if bodyLine == "" {
		t.Fatalf("rendered output did not contain mixed-direction text: %q", output)
	}
	if strings.HasPrefix(bodyLine, " ") {
		t.Fatalf("English-first mixed body was unexpectedly right-aligned: %q", bodyLine)
	}
}

func plainStyledRunes(tokens []ansiStyledRune) string {
	var out strings.Builder
	for _, token := range tokens {
		out.WriteRune(token.r)
	}
	return out.String()
}

func findLineContaining(text, fragment string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, fragment) {
			return line
		}
	}
	return ""
}
