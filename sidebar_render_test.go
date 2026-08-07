package main

import (
	"fmt"
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

func TestSelectedChatUsesOneContinuousFullWidthStyle(t *testing.T) {
	name := "Selected conversation"
	app := NewApp()
	app.Chats = []Chat{{ID: "selected", ChatType: "group", CachedDisplayName: &name}}
	app.SelectedIndex = 0
	model := NewModel(app, "client", "user")

	const width = 42
	lines := strings.Split(model.renderChatList(width, 8), "\n")
	if len(lines) != 2 {
		t.Fatalf("sidebar rendered %d lines, want one header and one row", len(lines))
	}
	selected := lines[1]
	if got := lipgloss.Width(selected); got != width {
		t.Fatalf("selected row width = %d, want %d", got, width)
	}
	if !strings.Contains(stripANSI(selected), "› ") {
		t.Fatalf("selected row omitted its cursor: %q", stripANSI(selected))
	}
	want := lipgloss.NewStyle().
		Foreground(colWhite).
		Background(colSelected).
		Bold(true).
		Width(width).
		MaxWidth(width).
		Render(stripANSI(selected))
	if selected != want {
		t.Fatalf("selected row was not rendered as one continuous full-width style: %q", selected)
	}
	if resets := strings.Count(selected, "\x1b[0m"); resets > 1 {
		t.Fatalf("selected row contains %d full style resets, so its background is not continuous: %q", resets, selected)
	}
}

func TestFullViewRendersFilteredChatHeaderOnce(t *testing.T) {
	app := NewApp()
	visible := make([]Chat, 36)
	for index := range visible {
		name := fmt.Sprintf("Visible chat %d", index)
		visible[index] = Chat{ID: fmt.Sprintf("visible-%d", index), CachedDisplayName: &name}
	}
	app.SetChats(visible)
	app.ActiveChatBookmark = "Unread"
	model := NewModel(app, "client", "user")
	for index := 0; index < 196; index++ {
		model.chatCache[fmt.Sprintf("chat-%d", index)] = Chat{ID: fmt.Sprintf("chat-%d", index)}
	}
	model.width = 120
	model.height = 30

	const header = "Chats 36/196 · Unread"
	if count := strings.Count(stripANSI(model.View()), header); count != 1 {
		t.Fatalf("full view rendered filtered chat header %d times, want once", count)
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
