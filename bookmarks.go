package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type chatBookmarkPreset struct {
	Key    string
	Name   string
	Filter ChatListFilter
}

func builtinChatBookmarkPresets() []chatBookmarkPreset {
	preset := func(key, name string, configure func(*ChatListFilter)) chatBookmarkPreset {
		filter := newChatListFilter()
		if configure != nil {
			configure(&filter)
		}
		return chatBookmarkPreset{Key: key, Name: name, Filter: filter}
	}

	return []chatBookmarkPreset{
		preset("i", "Inbox (all chats)", nil),
		preset("u", "Unread", func(filter *ChatListFilter) { filter.ReadState = ChatReadUnread }),
		preset("r", "Read", func(filter *ChatListFilter) { filter.ReadState = ChatReadRead }),
		preset("t", "Today", func(filter *ChatListFilter) { filter.TodayOnly = true }),
		preset("f", "Favorites", func(filter *ChatListFilter) { filter.FavouritesOnly = true }),
		preset("d", "Direct (1:1)", func(filter *ChatListFilter) { filter.ChatTypes["oneOnOne"] = true }),
		preset("g", "Groups", func(filter *ChatListFilter) { filter.ChatTypes["group"] = true }),
		preset("m", "Meetings", func(filter *ChatListFilter) { filter.ChatTypes["meeting"] = true }),
		preset("a", "All chats", nil),
	}
}

func bookmarkFilter(config ChatBookmarkConfig) ChatListFilter {
	filter := newChatListFilter()
	filter.Query = config.Query
	filter.ReadState = config.ReadState
	filter.FavouritesOnly = config.FavouritesOnly
	filter.TodayOnly = config.TodayOnly
	for _, chatType := range config.ChatTypes {
		filter.ChatTypes[chatType] = true
	}
	return filter
}

func (m Model) chatBookmarkPresets() []chatBookmarkPreset {
	presets := builtinChatBookmarkPresets()
	for _, config := range m.app.ChatBookmarks {
		custom := chatBookmarkPreset{Key: config.Key, Name: config.Name, Filter: bookmarkFilter(config)}
		replaced := false
		for index := range presets {
			if presets[index].Key == custom.Key {
				presets[index] = custom
				replaced = true
				break
			}
		}
		if !replaced {
			presets = append(presets, custom)
		}
	}
	return presets
}

func (m Model) applyChatBookmark(preset chatBookmarkPreset) (Model, tea.Cmd) {
	m.app.ChatBookmarkPopupMode = false
	// Inbox/All are explicit resets. The dedicated Unread bookmark remains a
	// standalone view, while every other bookmark retains the independent U
	// overlay so combinations such as Today + Unread stay composable.
	if preset.Key == "i" || preset.Key == "a" || preset.Key == "u" || preset.Filter.ReadState == ChatReadUnread {
		m.app.UnreadOverlay = false
	}
	m.app.DraftChatFilter = cloneChatListFilter(preset.Filter)
	updated, cmd := m.applyChatFilter()
	updated.app.ActiveChatBookmark = preset.Name
	updated.app.SetStatus(fmt.Sprintf("Bookmark %s: %d shown", preset.Name, len(updated.app.Chats)), 4*time.Second)
	return updated, cmd
}

func (m Model) handleChatBookmarkPopupKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	presets := m.chatBookmarkPresets()
	key := m.keyName(keyContextBookmarks, msg)
	switch key {
	case "esc", "q", "b":
		m.app.ChatBookmarkPopupMode = false
		return m, nil
	case "j", "down", "tab":
		m.app.ChatBookmarkSelectedIndex = (m.app.ChatBookmarkSelectedIndex + 1) % len(presets)
		return m, nil
	case "k", "up", "shift+tab":
		m.app.ChatBookmarkSelectedIndex--
		if m.app.ChatBookmarkSelectedIndex < 0 {
			m.app.ChatBookmarkSelectedIndex = len(presets) - 1
		}
		return m, nil
	case "enter":
		index := m.app.ChatBookmarkSelectedIndex
		if index >= 0 && index < len(presets) {
			return m.applyChatBookmark(presets[index])
		}
		return m, nil
	}

	for _, preset := range presets {
		if key == preset.Key {
			return m.applyChatBookmark(preset)
		}
	}
	return m, nil
}

func (m Model) renderChatBookmarkPopup(w, h int) string {
	if w < 42 {
		w = 42
	}
	presets := m.chatBookmarkPresets()
	lines := []string{
		lipgloss.NewStyle().Foreground(colYellow).Bold(true).Render("Chat bookmarks"),
		lipgloss.NewStyle().Foreground(colDimGray).Render("Choose a preset shortcut or select from the list"),
		"",
	}
	for index, preset := range presets {
		cursor := "  "
		style := lipgloss.NewStyle()
		if index == m.app.ChatBookmarkSelectedIndex {
			cursor = "› "
			style = style.Foreground(colCyan).Bold(true)
		}
		lines = append(lines, style.Render(fmt.Sprintf("%s%s  %s", cursor, preset.Key, preset.Name)))
	}
	lines = append(lines, "", lipgloss.NewStyle().Foreground(colDimGray).Render(
		m.keybindings.Display(keyListSelect)+" apply · "+m.keybindings.Display(keyBookmarkClose)+" cancel"))

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colGreen).
		Padding(1, 2).
		Width(w - 2).
		MaxHeight(h).
		Render(strings.Join(lines, "\n"))
}
