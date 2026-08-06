package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type testCSIMessage []byte

func newComposeKeyTestModel(value string) Model {
	app := NewApp()
	app.Chats = []Chat{{ID: "chat-1"}}
	app.SelectedIndex = 0
	app.InputMode = true
	model := NewModel(app, "client", "user")
	model.textarea.SetValue(value)
	model.textarea.CursorEnd()
	model.textarea.Focus()
	return model
}

func TestComposeRendersHebrewVisuallyWithoutChangingLogicalText(t *testing.T) {
	logical := "שלום עולם"
	model := newComposeKeyTestModel(logical)
	model.app.MessagesConversationID = "chat-1"

	rendered := stripANSI(model.renderRightPanel(60, 15))
	if !strings.Contains(rendered, "םלוע םולש") {
		t.Fatalf("compose view did not apply RTL visual ordering: %q", rendered)
	}
	if got := model.textarea.Value(); got != logical {
		t.Fatalf("compose rendering changed logical text to %q, want %q", got, logical)
	}
}

func TestComposeRendersHebrewInsideEnglishSentence(t *testing.T) {
	logical := "Hello שלום world"
	model := newComposeKeyTestModel(logical)
	model.app.MessagesConversationID = "chat-1"

	rendered := stripANSI(model.renderRightPanel(60, 15))
	if !strings.Contains(rendered, "Hello םולש world") {
		t.Fatalf("compose view did not reorder the Hebrew run in mixed text: %q", rendered)
	}
	if got := model.textarea.Value(); got != logical {
		t.Fatalf("compose rendering changed mixed logical text to %q", got)
	}
}

func TestComposeEnterInsertsOneNewline(t *testing.T) {
	model := newComposeKeyTestModel("first line")
	model, command := model.updateInternal(tea.KeyMsg{Type: tea.KeyEnter})

	if command != nil {
		t.Fatal("plain Enter unexpectedly returned a send command")
	}
	if !model.app.InputMode {
		t.Fatal("plain Enter left compose mode")
	}
	if got, want := model.textarea.Value(), "first line\n"; got != want {
		t.Fatalf("plain Enter produced %q, want %q", got, want)
	}
	if model.app.InputBuffer != model.textarea.Value() {
		t.Fatalf("input buffer %q does not match textarea %q", model.app.InputBuffer, model.textarea.Value())
	}
}

func TestComposeCtrlJSends(t *testing.T) {
	model := newComposeKeyTestModel("send this")
	model, command := model.updateInternal(tea.KeyMsg{Type: tea.KeyCtrlJ})

	if command == nil {
		t.Fatal("Ctrl+J did not return a send command")
	}
	if model.app.InputMode {
		t.Fatal("Ctrl+J did not leave compose mode")
	}
}

func TestComposeEnhancedCtrlEnterSends(t *testing.T) {
	for _, sequence := range []string{"\x1b[13;5u", "\x1b[27;5;13~"} {
		model := newComposeKeyTestModel("send this")
		model, command := model.updateInternal(testCSIMessage(sequence))

		if command == nil {
			t.Fatalf("enhanced Ctrl+Enter %q did not return a send command", sequence)
		}
		if model.app.InputMode {
			t.Fatalf("enhanced Ctrl+Enter %q did not leave compose mode", sequence)
		}
	}
}
