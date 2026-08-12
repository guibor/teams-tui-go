package main

import (
	"encoding/json"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestKeyListAcceptsStringOrArray(t *testing.T) {
	var config struct {
		Bindings KeyBindingConfig `json:"keybindings"`
	}
	if err := json.Unmarshal([]byte(`{"keybindings":{"search.global":"x","compose.send":["ctrl+s","C-s"]}}`), &config); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := config.Bindings[keySearchGlobal]; len(got) != 1 || got[0] != "x" {
		t.Fatalf("single binding = %#v", got)
	}
	if got := config.Bindings[keyComposeSend]; len(got) != 2 || got[1] != "C-s" {
		t.Fatalf("array binding = %#v", got)
	}
}

func TestDefaultKeyMapHasNoConflictsAndPreservesWorkflow(t *testing.T) {
	keyMap, warnings := NewKeyMap(nil)
	if len(warnings) != 0 {
		t.Fatalf("default bindings produced warnings: %v", warnings)
	}
	tests := []struct {
		context keyContext
		pressed string
		want    string
	}{
		{keyContextNormalChat, "s", "s"},
		{keyContextNormalChat, "N", "N"},
		{keyContextNormalChat, "U", "U"},
		{keyContextNormalChat, "M-n", "j"},
		{keyContextNormalChat, "h", "alt+<"},
		{keyContextNormalChannel, "h", "h"},
		{keyContextCompose, "Ctrl+Enter", "ctrl+j"},
		{keyContextCompose, "Enter", "enter"},
		{keyContextMessageSelect, "i", "r"},
		{keyContextNewChatResults, " ", "enter"},
	}
	for _, test := range tests {
		if got := keyMap.Canonical(test.context, test.pressed); got != test.want {
			t.Errorf("Canonical(%s, %q) = %q, want %q", test.context, test.pressed, got, test.want)
		}
	}
}

func TestKeyOverrideDisablesOldDefaultAndStaysModeScoped(t *testing.T) {
	keyMap, warnings := NewKeyMap(KeyBindingConfig{
		keySearchGlobal: KeyList{"x"},
		keyComposeSend:  KeyList{"ctrl+s", "C-s"},
	})
	if len(warnings) != 0 {
		t.Fatalf("override warnings: %v", warnings)
	}
	if got := keyMap.Canonical(keyContextNormalChat, "x"); got != "s" {
		t.Fatalf("custom search key = %q, want canonical s", got)
	}
	if got := keyMap.Canonical(keyContextNormalChat, "s"); got != "" {
		t.Fatalf("old search default remained active as %q", got)
	}
	if got := keyMap.Canonical(keyContextCompose, "ctrl+s"); got != "ctrl+j" {
		t.Fatalf("custom send key = %q, want canonical ctrl+j", got)
	}
	if got := keyMap.Canonical(keyContextNormalChat, "ctrl+s"); got != "ctrl+s" {
		t.Fatalf("compose-only key leaked into normal mode as %q", got)
	}
}

func TestEmptyKeyListUnbindsAction(t *testing.T) {
	keyMap, warnings := NewKeyMap(KeyBindingConfig{keyChatAnalyze: KeyList{}})
	if len(warnings) != 0 {
		t.Fatalf("unbind warnings: %v", warnings)
	}
	if got := keyMap.Canonical(keyContextNormalChat, "A"); got != "" {
		t.Fatalf("unbound action retained default %q", got)
	}
	if got := keyMap.Display(keyChatAnalyze); got != "unbound" {
		t.Fatalf("unbound display = %q", got)
	}
}

func TestKeyConflictAndUnknownActionWarnings(t *testing.T) {
	_, warnings := NewKeyMap(KeyBindingConfig{
		keySearchGlobal:  KeyList{"x"},
		keyNewChatOpen:   KeyList{"x"},
		"unknown.action": KeyList{"z"},
	})
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "assigned to both") {
		t.Fatalf("missing conflict warning: %v", warnings)
	}
	if !strings.Contains(joined, "unknown keybinding action") {
		t.Fatalf("missing unknown-action warning: %v", warnings)
	}
}

func TestConfiguredBindingWinsAConflictWithAnotherActionDefault(t *testing.T) {
	keyMap, warnings := NewKeyMap(KeyBindingConfig{keySearchGlobal: KeyList{"N"}})
	if len(warnings) == 0 {
		t.Fatal("expected a conflict warning")
	}
	if got := keyMap.Canonical(keyContextNormalChat, "N"); got != "s" {
		t.Fatalf("configured key resolved to %q, want search canonical s", got)
	}
}

func TestModelUsesConfiguredNormalBinding(t *testing.T) {
	model := newWorkflowTestModel()
	model.keybindings, _ = NewKeyMap(KeyBindingConfig{keyNewChatOpen: KeyList{"x"}})
	model, _ = model.handleNormalModeKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if !model.app.NewChatMode || !model.app.UserSearchPopupMode {
		t.Fatal("configured new-chat binding did not open the participant picker")
	}

	model = newWorkflowTestModel()
	model.keybindings, _ = NewKeyMap(KeyBindingConfig{keyNewChatOpen: KeyList{"x"}})
	model, _ = model.handleNormalModeKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	if model.app.NewChatMode {
		t.Fatal("overridden default N still opened the participant picker")
	}
}
