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

func chatBookmarkPresets() []chatBookmarkPreset {
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

func (m Model) applyChatBookmark(preset chatBookmarkPreset) (Model, tea.Cmd) {
	m.app.ChatBookmarkPopupMode = false
	m.app.DraftChatFilter = cloneChatListFilter(preset.Filter)
	updated, cmd := m.applyChatFilter()
	updated.app.SetStatus(fmt.Sprintf("Bookmark %s: %d shown", preset.Name, len(updated.app.Chats)), 4*time.Second)
	return updated, cmd
}

func (m Model) handleChatBookmarkPopupKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	presets := chatBookmarkPresets()
	switch msg.String() {
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
		if msg.String() == preset.Key {
			return m.applyChatBookmark(preset)
		}
	}
	return m, nil
}

func (m Model) renderChatBookmarkPopup(w, h int) string {
	if w < 42 {
		w = 42
	}
	presets := chatBookmarkPresets()
	lines := []string{
		lipgloss.NewStyle().Foreground(colYellow).Bold(true).Render("Chat bookmarks"),
		lipgloss.NewStyle().Foreground(colDimGray).Render("b followed by the highlighted shortcut"),
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
	lines = append(lines, "", lipgloss.NewStyle().Foreground(colDimGray).Render("Enter apply · Esc cancel"))

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colGreen).
		Padding(1, 2).
		Width(w - 2).
		MaxHeight(h).
		Render(strings.Join(lines, "\n"))
}
