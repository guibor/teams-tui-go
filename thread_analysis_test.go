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
		"gpt-5.6-luna",
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
		!strings.Contains(environment, "TEAMS_THREAD_ANALYSIS_MODEL=gpt-5.6-luna") {
		t.Fatalf("command environment omitted analysis selection: %s", environment)
	}
}

func TestAnalysisChooserSelectsDestinationThenModel(t *testing.T) {
	model := newWorkflowChatListModel("chat-1", "chat-2")
	model.app.ThreadAnalysisModels = []string{"gpt-5.6-sol", "gpt-5.6-luna"}
	model.app.ThreadAnalysisModel = "gpt-5.6-luna"
	model, _ = model.executeThreadAction(threadActionAnalyzeChoose)
	if !model.app.ThreadAnalysisPopupMode || model.app.ThreadAnalysisStage != 0 {
		t.Fatal("chooser did not open on destination stage")
	}
	model.app.ThreadAnalysisSelectedIndex = 0
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
