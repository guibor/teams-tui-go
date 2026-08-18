package main

import "testing"

func TestCustomBookmarksExtendAndOverrideBuiltins(t *testing.T) {
	app := NewApp()
	app.ChatBookmarks = []ChatBookmarkConfig{
		{Key: "u", Name: "Urgent unread", Query: "urgent", ReadState: ChatReadUnread},
		{Key: "p", Name: "Planning", Query: `in:"Product planning"`, ChatTypes: []string{"group"}},
	}
	model := NewModel(app, "client", "user")
	presets := model.chatBookmarkPresets()

	foundOverride := false
	foundCustom := false
	for _, preset := range presets {
		switch preset.Key {
		case "u":
			foundOverride = preset.Name == "Urgent unread" && preset.Filter.Query == "urgent" && preset.Filter.ReadState == ChatReadUnread
		case "p":
			foundCustom = preset.Name == "Planning" && preset.Filter.ChatTypes["group"]
		}
	}
	if !foundOverride || !foundCustom {
		t.Fatalf("custom bookmarks were not merged: %#v", presets)
	}
}

func TestRollingActivityBookmarks(t *testing.T) {
	presets := builtinChatBookmarkPresets()
	windows := map[string]int{"2": 24, "w": 7 * 24}
	for key, want := range windows {
		found := false
		for _, preset := range presets {
			if preset.Key == key {
				found = true
				if preset.Filter.WithinHours != want {
					t.Fatalf("bookmark b%s window = %d, want %d", key, preset.Filter.WithinHours, want)
				}
			}
		}
		if !found {
			t.Fatalf("bookmark b%s not found", key)
		}
	}
}

func TestResolveChatBookmarksValidatesEntries(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	InitConfig()
	cfg := LoadConfig()
	cfg.ChatBookmarks = []ChatBookmarkConfig{
		{Key: "p", Name: " Planning ", ReadState: ChatReadFilter("invalid"), ChatTypes: []string{"group", "group", ""}},
		{Key: "too-long", Name: "Ignored"},
		{Key: "x", Name: ""},
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	bookmarks := ResolveChatBookmarks()
	if len(bookmarks) != 1 {
		t.Fatalf("resolved bookmarks = %#v", bookmarks)
	}
	if bookmarks[0].Name != "Planning" || bookmarks[0].ReadState != ChatReadAll || len(bookmarks[0].ChatTypes) != 1 {
		t.Fatalf("bookmark was not normalized: %#v", bookmarks[0])
	}
}
