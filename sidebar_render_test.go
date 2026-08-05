package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestSelectedLongChatNameStaysOnOneSidebarRow(t *testing.T) {
	name := "Project Apollo coordination with architecture and infrastructure"
	app := NewApp()
	app.Chats = []Chat{{ID: "long-chat", ChatType: "group", CachedDisplayName: &name}}
	app.SelectedIndex = 0
	app.ShowChatDates = true
	model := NewModel(app, "client", "user")

	const width = 40
	rendered := model.renderChatList(width, 10)
	lines := strings.Split(rendered, "\n")
	if len(lines) != 2 {
		t.Fatalf("selected chat rendered as %d lines, want title plus one row:\n%s", len(lines), stripANSI(rendered))
	}
	if !strings.Contains(stripANSI(lines[1]), "…") {
		t.Fatalf("long chat name was not visibly truncated: %q", stripANSI(lines[1]))
	}
	if strings.Contains(stripANSI(lines[1]), name) {
		t.Fatalf("sidebar retained the complete overflowing name: %q", stripANSI(lines[1]))
	}
	if got := lipgloss.Width(lines[1]); got != width {
		t.Fatalf("selected row width = %d, want exactly %d cells", got, width)
	}
}

func TestRightPaneHeaderShowsCompleteSelectedChatName(t *testing.T) {
	name := "Project Apollo coordination with architecture and infrastructure"
	chat := Chat{ID: "long-chat", ChatType: "group", CachedDisplayName: &name}
	app := NewApp()
	app.Chats = []Chat{chat}
	app.SelectedIndex = 0
	app.MessagesConversationID = chat.ID
	model := NewModel(app, "client", "user")

	rendered := stripANSI(model.renderRightPanel(72, 12))
	nameAt := strings.Index(rendered, name)
	messagesAt := strings.Index(rendered, "No messages.")
	if nameAt < 0 {
		t.Fatalf("right pane omitted full chat name %q:\n%s", name, rendered)
	}
	if messagesAt < 0 || nameAt > messagesAt {
		t.Fatalf("full chat name was not rendered above the transcript:\n%s", rendered)
	}

	wrappedHeader := stripANSI(model.renderConversationHeader(24, ""))
	if got := strings.Join(strings.Fields(wrappedHeader), " "); got != name {
		t.Fatalf("wrapped right-pane header lost part of the full name: got %q, want %q", got, name)
	}
	if len(strings.Split(wrappedHeader, "\n")) < 2 {
		t.Fatalf("narrow right-pane header did not wrap the full name: %q", wrappedHeader)
	}
}
