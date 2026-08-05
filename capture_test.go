package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCaptureChatMarkdownGroupsByMarkedDayAndDeduplicates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes", "teams-threads.md")
	title := "Design [review]"
	chat := Chat{
		ID:                "chat-1",
		CachedDisplayName: &title,
		WebURL:            "https://teams.microsoft.com/l/chat/example",
	}
	dayOne := time.Date(2026, time.August, 2, 9, 30, 0, 0, time.FixedZone("IDT", 3*60*60))

	resolved, added, err := CaptureChatMarkdown(path, chat, dayOne)
	if err != nil {
		t.Fatalf("CaptureChatMarkdown failed: %v", err)
	}
	if !added || resolved != path {
		t.Fatalf("first capture returned path=%q added=%t", resolved, added)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("capture file missing: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("capture mode is %o, want 600", info.Mode().Perm())
	}

	if _, added, err = CaptureChatMarkdown(path, chat, dayOne.Add(time.Hour)); err != nil || added {
		t.Fatalf("same-day duplicate returned added=%t err=%v", added, err)
	}
	dayTwo := dayOne.Add(24 * time.Hour)
	if _, added, err = CaptureChatMarkdown(path, chat, dayTwo); err != nil || !added {
		t.Fatalf("next-day capture returned added=%t err=%v", added, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	contents := string(data)
	for _, want := range []string{
		"# Teams Threads",
		"## 2026-08-02",
		"## 2026-08-03",
		"- [ ] [Design \\[review\\]](<https://teams.microsoft.com/l/chat/example>)",
	} {
		if !strings.Contains(contents, want) {
			t.Errorf("capture missing %q:\n%s", want, contents)
		}
	}
	if count := strings.Count(contents, captureMarker(chat)); count != 2 {
		t.Fatalf("marker appears %d times, want once per day", count)
	}
}

func TestCaptureChatMarkdownInsertsIntoExistingDay(t *testing.T) {
	contents := "# Teams Threads\n\n## 2026-08-02\n\n- [ ] Existing\n\n## 2026-08-01\n\n- [ ] Older\n"
	title := "New chat"
	chat := Chat{ID: "new", CachedDisplayName: &title}
	markedAt := time.Date(2026, time.August, 2, 20, 0, 0, 0, time.Local)

	updated, added := insertCaptureEntry(contents, chat, markedAt)
	if !added {
		t.Fatal("entry was not added")
	}
	newIndex := strings.Index(updated, "New chat")
	existingIndex := strings.Index(updated, "Existing")
	olderHeadingIndex := strings.Index(updated, "## 2026-08-01")
	if newIndex < 0 || !(newIndex < existingIndex && existingIndex < olderHeadingIndex) {
		t.Fatalf("entry was not inserted in the matching day:\n%s", updated)
	}
}

func TestCaptureChatOrgGroupsByMarkedDayAndDeduplicates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes", "teams-threads.org")
	title := "Design [review]"
	chat := Chat{
		ID:                "chat-1",
		CachedDisplayName: &title,
		WebURL:            "https://teams.microsoft.com/l/chat/example",
	}
	dayOne := time.Date(2026, time.August, 2, 9, 30, 0, 0, time.Local)

	resolved, added, err := CaptureChat(path, ThreadCaptureOrg, chat, dayOne)
	if err != nil {
		t.Fatalf("CaptureChat Org failed: %v", err)
	}
	if !added || resolved != path {
		t.Fatalf("first Org capture returned path=%q added=%t", resolved, added)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Org capture file missing: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("Org capture mode is %o, want 600", info.Mode().Perm())
	}

	if _, added, err = CaptureChat(path, ThreadCaptureOrg, chat, dayOne.Add(time.Hour)); err != nil || added {
		t.Fatalf("same-day Org duplicate returned added=%t err=%v", added, err)
	}
	dayTwo := dayOne.Add(24 * time.Hour)
	if _, added, err = CaptureChat(path, ThreadCaptureOrg, chat, dayTwo); err != nil || !added {
		t.Fatalf("next-day Org capture returned added=%t err=%v", added, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	contents := string(data)
	for _, want := range []string{
		"#+title: Teams Threads",
		"* 2026-08-02",
		"* 2026-08-03",
		"** TODO Teams: Design [review]",
		":CAPTURED: [2026-08-02 Sun 09:30]",
		":TEAMS_KIND: chat",
		":TEAMS_CHAT: chat-1",
		":TEAMS_URL: https://teams.microsoft.com/l/chat/example",
		"[[https://teams.microsoft.com/l/chat/example][Open in Microsoft Teams]]",
	} {
		if !strings.Contains(contents, want) {
			t.Errorf("Org capture missing %q:\n%s", want, contents)
		}
	}
	if count := strings.Count(contents, orgCaptureMarker(chat)); count != 2 {
		t.Fatalf("Org marker appears %d times, want once per day", count)
	}
}

func TestCaptureChatOrgInsertsIntoExistingDay(t *testing.T) {
	contents := "#+title: Teams Threads\n\n* 2026-08-02\n\n** TODO Teams: Existing\n\n* 2026-08-01\n\n** TODO Teams: Older\n"
	title := "New chat"
	chat := Chat{ID: "new", CachedDisplayName: &title}
	markedAt := time.Date(2026, time.August, 2, 20, 0, 0, 0, time.Local)

	updated, added := insertOrgCaptureEntry(contents, chat, markedAt)
	if !added {
		t.Fatal("Org entry was not added")
	}
	newIndex := strings.Index(updated, "New chat")
	existingIndex := strings.Index(updated, "Existing")
	olderHeadingIndex := strings.Index(updated, "* 2026-08-01")
	if newIndex < 0 || !(newIndex < existingIndex && existingIndex < olderHeadingIndex) {
		t.Fatalf("Org entry was not inserted in the matching day:\n%s", updated)
	}
}
