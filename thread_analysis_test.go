package main

import (
	"reflect"
	"testing"
)

func TestBuildThreadAnalysisCommandKeepsExportPathAsOneArgument(t *testing.T) {
	cmd, err := buildThreadAnalysisCommand(
		"/usr/local/bin/mdf-teams-agent-shell --server spacemacs",
		"claude",
		"/tmp/Team thread with spaces.md",
	)
	if err != nil {
		t.Fatalf("buildThreadAnalysisCommand failed: %v", err)
	}
	want := []string{
		"/usr/local/bin/mdf-teams-agent-shell",
		"--server",
		"spacemacs",
		"--agent",
		"claude",
		"/tmp/Team thread with spaces.md",
	}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("command args = %#v, want %#v", cmd.Args, want)
	}
}

func TestBuildThreadAnalysisCommandRejectsEmptyCommand(t *testing.T) {
	if _, err := buildThreadAnalysisCommand("  ", "codex", "/tmp/thread.md"); err == nil {
		t.Fatal("expected empty thread analysis command to fail")
	}
}
