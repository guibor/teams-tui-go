package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func expandCaptureFile(path, fallback string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = fallback
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return filepath.Clean(path), nil
}

func captureIdentity(chat Chat) string {
	identity := strings.TrimSpace(chat.ID)
	if identity == "" {
		identity = strings.TrimSpace(chat.WebURL)
	}
	return identity
}

func encodedCaptureIdentity(chat Chat) string {
	return base64.RawURLEncoding.EncodeToString([]byte(captureIdentity(chat)))
}

func markdownLinkLabel(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	return strings.NewReplacer(
		`\`, `\\`,
		`[`, `\[`,
		`]`, `\]`,
	).Replace(value)
}

func captureMarker(chat Chat) string {
	return "<!-- teams-chat:" + encodedCaptureIdentity(chat) + " -->"
}

func captureEntry(chat Chat) string {
	name := markdownLinkLabel(chatExportTitle(chat))
	url := strings.TrimSpace(chat.WebURL)
	if url == "" {
		return fmt.Sprintf("- [ ] %s %s", name, captureMarker(chat))
	}
	url = strings.ReplaceAll(url, ">", "%3E")
	return fmt.Sprintf("- [ ] [%s](<%s>) %s", name, url, captureMarker(chat))
}

func orgCaptureMarker(chat Chat) string {
	return ":TEAMS_CAPTURE_KEY: " + encodedCaptureIdentity(chat)
}

func orgOneLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func orgLinkTarget(value string) string {
	value = strings.TrimSpace(value)
	return strings.NewReplacer(
		"[", "%5B",
		"]", "%5D",
		" ", "%20",
	).Replace(value)
}

func orgCaptureEntry(chat Chat, markedAt time.Time) string {
	title := orgOneLine(chatExportTitle(chat))
	url := orgOneLine(chat.WebURL)
	lines := []string{
		"** TODO Teams: " + title,
		":PROPERTIES:",
		":CAPTURED: [" + markedAt.Local().Format("2006-01-02 Mon 15:04") + "]",
		":TEAMS_KIND: chat",
		":TEAMS_TITLE: " + title,
	}
	if chatID := orgOneLine(chat.ID); chatID != "" {
		lines = append(lines, ":TEAMS_CHAT: "+chatID)
	}
	if url != "" {
		lines = append(lines, ":TEAMS_URL: "+url)
	}
	lines = append(lines, orgCaptureMarker(chat), ":END:")
	if url != "" {
		lines = append(lines, "", "[["+orgLinkTarget(url)+"][Open in Microsoft Teams]]")
	}
	return strings.Join(lines, "\n")
}

func insertLines(lines []string, index int, values []string) []string {
	result := make([]string, 0, len(lines)+len(values))
	result = append(result, lines[:index]...)
	result = append(result, values...)
	result = append(result, lines[index:]...)
	return result
}

func insertCaptureEntry(contents string, chat Chat, markedAt time.Time) (string, bool) {
	contents = strings.ReplaceAll(contents, "\r\n", "\n")
	contents = strings.TrimRight(contents, "\n")
	if strings.TrimSpace(contents) == "" {
		contents = "# Teams Threads"
	}
	lines := strings.Split(contents, "\n")
	heading := "## " + markedAt.Local().Format("2006-01-02")
	marker := captureMarker(chat)
	start := -1
	end := len(lines)
	for index, line := range lines {
		if strings.TrimSpace(line) == heading {
			start = index
			continue
		}
		if start >= 0 && strings.HasPrefix(strings.TrimSpace(line), "## ") {
			end = index
			break
		}
	}
	if start >= 0 {
		for _, line := range lines[start+1 : end] {
			if strings.Contains(line, marker) {
				return contents + "\n", false
			}
		}
		insertAt := start + 1
		if insertAt < len(lines) && strings.TrimSpace(lines[insertAt]) == "" {
			insertAt++
			lines = append(lines[:insertAt], append([]string{captureEntry(chat)}, lines[insertAt:]...)...)
		} else {
			lines = append(lines[:insertAt], append([]string{"", captureEntry(chat)}, lines[insertAt:]...)...)
		}
		return strings.Join(lines, "\n") + "\n", true
	}

	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
		lines = append(lines, "")
	}
	lines = append(lines, heading, "", captureEntry(chat))
	return strings.Join(lines, "\n") + "\n", true
}

func insertOrgCaptureEntry(contents string, chat Chat, markedAt time.Time) (string, bool) {
	contents = strings.ReplaceAll(contents, "\r\n", "\n")
	contents = strings.TrimRight(contents, "\n")
	if strings.TrimSpace(contents) == "" {
		contents = "#+title: Teams Threads"
	}
	lines := strings.Split(contents, "\n")
	heading := "* " + markedAt.Local().Format("2006-01-02")
	marker := orgCaptureMarker(chat)
	start := -1
	end := len(lines)
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == heading {
			start = index
			continue
		}
		if start >= 0 && strings.HasPrefix(trimmed, "* ") {
			end = index
			break
		}
	}
	if start >= 0 {
		for _, line := range lines[start+1 : end] {
			if strings.TrimSpace(line) == marker {
				return contents + "\n", false
			}
		}

		entryLines := strings.Split(orgCaptureEntry(chat, markedAt), "\n")
		insertAt := start + 1
		if insertAt < len(lines) && strings.TrimSpace(lines[insertAt]) == "" {
			insertAt++
			entryLines = append(entryLines, "")
		} else {
			entryLines = append([]string{""}, entryLines...)
			entryLines = append(entryLines, "")
		}
		lines = insertLines(lines, insertAt, entryLines)
		return strings.Join(lines, "\n") + "\n", true
	}

	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
		lines = append(lines, "")
	}
	lines = append(lines, heading, "")
	lines = append(lines, strings.Split(orgCaptureEntry(chat, markedAt), "\n")...)
	return strings.Join(lines, "\n") + "\n", true
}

type captureInserter func(string, Chat, time.Time) (string, bool)

func captureChatFile(
	path string,
	fallback string,
	temporaryPattern string,
	chat Chat,
	markedAt time.Time,
	insert captureInserter,
) (string, bool, error) {
	path, err := expandCaptureFile(path, fallback)
	if err != nil {
		return "", false, fmt.Errorf("resolve capture file: %w", err)
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", false, fmt.Errorf("create capture directory: %w", err)
	}

	contents := ""
	if data, readErr := os.ReadFile(path); readErr == nil {
		contents = string(data)
	} else if !os.IsNotExist(readErr) {
		return "", false, fmt.Errorf("read capture file: %w", readErr)
	}
	updated, added := insert(contents, chat, markedAt)
	if !added {
		return path, false, nil
	}

	temporary, err := os.CreateTemp(directory, temporaryPattern)
	if err != nil {
		return "", false, fmt.Errorf("create temporary capture file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return "", false, fmt.Errorf("secure temporary capture file: %w", err)
	}
	if _, err := temporary.WriteString(updated); err != nil {
		temporary.Close()
		return "", false, fmt.Errorf("write capture file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", false, fmt.Errorf("close capture file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return "", false, fmt.Errorf("replace capture file: %w", err)
	}
	return path, true, nil
}

// CaptureChat records a thread in the configured capture syntax.
func CaptureChat(path string, format ThreadCaptureFormat, chat Chat, markedAt time.Time) (string, bool, error) {
	if normalizeThreadCaptureFormat(format) == ThreadCaptureOrg {
		return CaptureChatOrg(path, chat, markedAt)
	}
	return CaptureChatMarkdown(path, chat, markedAt)
}

// CaptureChatMarkdown records a chat once per local day in a private Markdown checklist.
func CaptureChatMarkdown(path string, chat Chat, markedAt time.Time) (string, bool, error) {
	return captureChatFile(
		path,
		"~/Documents/teams-threads.md",
		".teams-threads-*.md",
		chat,
		markedAt,
		insertCaptureEntry,
	)
}

// CaptureChatOrg records a chat once per local day as an Org TODO capture.
func CaptureChatOrg(path string, chat Chat, markedAt time.Time) (string, bool, error) {
	return captureChatFile(
		path,
		"~/Documents/teams-threads.org",
		".teams-threads-*.org",
		chat,
		markedAt,
		insertOrgCaptureEntry,
	)
}
