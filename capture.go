package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func expandCaptureFile(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "~/Documents/teams-threads.md"
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

func markdownLinkLabel(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	return strings.NewReplacer(
		`\`, `\\`,
		`[`, `\[`,
		`]`, `\]`,
	).Replace(value)
}

func captureMarker(chat Chat) string {
	identity := chat.ID
	if identity == "" {
		identity = strings.TrimSpace(chat.WebURL)
	}
	encoded := base64.RawURLEncoding.EncodeToString([]byte(identity))
	return "<!-- teams-chat:" + encoded + " -->"
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

// CaptureChatMarkdown records a chat once per local day in a private Markdown checklist.
func CaptureChatMarkdown(path string, chat Chat, markedAt time.Time) (string, bool, error) {
	path, err := expandCaptureFile(path)
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
	updated, added := insertCaptureEntry(contents, chat, markedAt)
	if !added {
		return path, false, nil
	}

	temporary, err := os.CreateTemp(directory, ".teams-threads-*.md")
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
