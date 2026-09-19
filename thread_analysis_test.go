package main

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBuildThreadAnalysisCommandKeepsExportPathAsOneArgument(t *testing.T) {
	cmd, err := buildThreadAnalysisCommand(
		"/usr/local/bin/thread-analysis-bridge --profile default",
		"claude",
		"terminal",
		"example-model",
		"/tmp/Team thread with spaces.md",
	)
	if err != nil {
		t.Fatalf("buildThreadAnalysisCommand failed: %v", err)
	}
	want := []string{
		"/usr/local/bin/thread-analysis-bridge",
		"--profile",
		"default",
		"--agent",
		"claude",
		"/tmp/Team thread with spaces.md",
	}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("command args = %#v, want %#v", cmd.Args, want)
	}
	environment := strings.Join(cmd.Env, "\n")
	if !strings.Contains(environment, "TEAMS_THREAD_ANALYSIS_DESTINATION=terminal") ||
		!strings.Contains(environment, "TEAMS_THREAD_ANALYSIS_MODEL=example-model") {
		t.Fatalf("command environment omitted analysis selection: %s", environment)
	}
}

func TestAnalysisChooserSelectsDestinationThenModel(t *testing.T) {
	model := newWorkflowChatListModel("chat-1", "chat-2")
	model.app.ThreadAnalysisDestination = "codex-app"
	model.app.ThreadAnalysisModels = []string{"other-model", "example-model"}
	model.app.ThreadAnalysisModel = "example-model"
	model, _ = model.executeThreadAction(threadActionAnalyzeChoose)
	if !model.app.ThreadAnalysisPopupMode || model.app.ThreadAnalysisStage != 0 {
		t.Fatal("chooser did not open on destination stage")
	}
	if model.app.ThreadAnalysisSelectedIndex != 2 {
		t.Fatalf("destination selected=%d, want remembered codex-app", model.app.ThreadAnalysisSelectedIndex)
	}
	model, _ = model.handleThreadAnalysisPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if model.app.ThreadAnalysisStage != 1 || model.app.ThreadAnalysisSelectedIndex != 1 {
		t.Fatalf("model stage=%d selected=%d, want stage 1 selected configured model", model.app.ThreadAnalysisStage, model.app.ThreadAnalysisSelectedIndex)
	}
}

func TestBuildThreadAnalysisCommandRejectsEmptyCommand(t *testing.T) {
	if _, err := buildThreadAnalysisCommand("  ", "codex", "terminal", "default", "/tmp/thread.md"); err == nil {
		t.Fatal("expected empty thread analysis command to fail")
	}
}

func TestAnalysisChooserUsesCustomDestinations(t *testing.T) {
	m := newWorkflowChatListModel("chat-1")
	m.app.ThreadAnalysisDestinations = []string{"local-tool", "remote-tool"}
	m.app.ThreadAnalysisDestination = "remote-tool"
	m.app.ThreadAnalysisModels = []string{"default"}
	m, _ = m.executeThreadAction(threadActionAnalyzeChoose)
	if !reflect.DeepEqual(m.threadAnalysisChoices(), []string{"local-tool", "remote-tool"}) || m.app.ThreadAnalysisSelectedIndex != 1 {
		t.Fatal("chooser did not use custom routes and remembered selection")
	}
	m, _ = m.handleThreadAnalysisPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if m.app.ThreadAnalysisPendingDestination != "remote-tool" || m.app.ThreadAnalysisStage != 1 {
		t.Fatal("custom route was not retained for model selection")
	}
}

func TestBuildThreadAnalysisCommandExpandsRoutingPlaceholders(t *testing.T) {
	cmd, err := buildThreadAnalysisCommand(
		"/bridge --destination {destination} --model={model}",
		"codex", "codex-app", "example-model", "/tmp/thread.md")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/bridge", "--destination", "codex-app", "--model=example-model", "--agent", "codex", "/tmp/thread.md"}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("command args = %#v, want %#v", cmd.Args, want)
	}
}
