package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMetaAngleKeysJumpToChatListEnds(t *testing.T) {
	app := NewApp()
	app.Chats = []Chat{{ID: "first"}, {ID: "middle"}, {ID: "last"}}
	app.SelectedIndex = 1
	model := NewModel(app, "client", "user")

	firstKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'<'}, Alt: true}
	if firstKey.String() != "alt+<" {
		t.Fatalf("unexpected Bubble Tea Meta-< encoding: %q", firstKey.String())
	}
	model, _ = model.handleNormalModeKey(firstKey)
	if model.app.SelectedIndex != 0 {
		t.Fatalf("Meta-< selected index %d, want 0", model.app.SelectedIndex)
	}

	lastKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'>'}, Alt: true}
	if lastKey.String() != "alt+>" {
		t.Fatalf("unexpected Bubble Tea Meta-> encoding: %q", lastKey.String())
	}
	model, _ = model.handleNormalModeKey(lastKey)
	if model.app.SelectedIndex != len(model.app.Chats)-1 {
		t.Fatalf("Meta-> selected index %d, want %d", model.app.SelectedIndex, len(model.app.Chats)-1)
	}
}
