package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestQLeavesConversationBeforeQuitting(t *testing.T) {
	app := NewApp()
	app.Chats = []Chat{{ID: "chat-1"}}
	app.SelectedIndex = 0
	app.Messages = []Message{{ID: "message-1"}}
	app.NextLink = "next"
	app.MessageSelectionMode = false
	model := NewModel(app, "client", "user")

	model, cmd := model.handleNormalModeKey(filterTestKey('q'))
	if cmd != nil {
		t.Fatal("first q returned a quit command")
	}
	if app.SelectedIndex != -1 || len(app.Messages) != 0 || app.NextLink != "" {
		t.Fatalf("first q did not clear the active conversation: index=%d messages=%d next=%q", app.SelectedIndex, len(app.Messages), app.NextLink)
	}
	if got := model.renderRightPanel(70, 20); !strings.Contains(got, "Dashboard") || strings.Contains(got, "Sleep Mode") {
		t.Fatalf("first q did not render the dashboard:\n%s", got)
	}

	_, cmd = model.handleNormalModeKey(filterTestKey('q'))
	if cmd == nil {
		t.Fatal("second q did not return a quit command")
	}
	quitMsg := cmd()
	if _, ok := quitMsg.(tea.QuitMsg); !ok {
		t.Fatalf("second q returned %T, want tea.QuitMsg", quitMsg)
	}
}

func TestCtrlCQuitsImmediatelyFromConversation(t *testing.T) {
	app := NewApp()
	app.Chats = []Chat{{ID: "chat-1"}}
	app.SelectedIndex = 0
	model := NewModel(app, "client", "user")

	_, cmd := model.handleNormalModeKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("Ctrl+C did not return a quit command")
	}
	quitMsg := cmd()
	if _, ok := quitMsg.(tea.QuitMsg); !ok {
		t.Fatalf("Ctrl+C returned %T, want tea.QuitMsg", quitMsg)
	}
}
