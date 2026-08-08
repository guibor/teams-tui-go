package main

import "testing"

func TestBuildExternalEditorCommandWithArguments(t *testing.T) {
	command, err := buildExternalEditorCommand(
		"/usr/local/bin/emacsclient --wait",
		"/tmp/teams message.txt",
	)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"/usr/local/bin/emacsclient",
		"--wait",
		"/tmp/teams message.txt",
	}
	if len(command.Args) != len(want) {
		t.Fatalf("command args = %#v, want %#v", command.Args, want)
	}
	for index := range want {
		if command.Args[index] != want[index] {
			t.Fatalf("command arg %d = %q, want %q", index, command.Args[index], want[index])
		}
	}
}

func TestBuildExternalEditorCommandRejectsEmptyCommand(t *testing.T) {
	if _, err := buildExternalEditorCommand("  ", "/tmp/message.txt"); err == nil {
		t.Fatal("expected an empty editor command to fail")
	}
}
