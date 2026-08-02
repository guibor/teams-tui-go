package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) hasActiveConversation() bool {
	return m.app.GetSelectedChat() != nil || m.channelSelectedIndex >= 0 ||
		m.app.SelectedChannelID != "" || m.app.SelectedChannelTeamID != ""
}

func (m Model) leaveActiveConversation() Model {
	m.app.SelectedIndex = -1
	m.channelSelectedIndex = -1
	m.app.SelectedChannelTeamID = ""
	m.app.SelectedChannelID = ""
	m.app.ChannelReplyToID = ""
	m.app.Messages = nil
	m.app.NextLink = ""
	m.app.PendingScrollID = ""
	m.app.ScrollOffset = 0
	m.app.MaxScroll = 0
	m.app.SnapToBottom = true
	m.app.SetLoadingMessages(false)
	m.app.MessageSelectionMode = false
	m.app.MessagePopupMode = false
	m.app.MessageSelectedIndex = 0
	m.app.ReplyToMessage = nil
	m.app.SearchMode = false
	m.app.SearchActive = false
	m.app.SearchQuery = ""
	m.app.SearchPopupMode = false
	m.app.SearchPopupResults = nil
	m.app.UrlSelectionMode = false
	return m
}

func (m Model) renderDashboard(w, h int) string {
	unread := 0
	for _, chat := range m.app.Chats {
		if m.isUnread(chat) {
			unread++
		}
	}
	lines := []string{
		lipgloss.NewStyle().Foreground(colWhite).Bold(true).Render("Teams"),
		"",
		lipgloss.NewStyle().Foreground(colDimGray).Render(
			fmt.Sprintf("Unread %d   Favorites %d   Visible %d", unread, len(m.favourites), len(m.app.Chats)),
		),
	}
	if chatFilterIsActive(m.app.ActiveChatFilter) {
		lines = append(lines, lipgloss.NewStyle().Foreground(colDimGray).Render("Filter  "+chatFilterSummary(m.app.ActiveChatFilter)))
	}
	body := lipgloss.Place(w, h-2, lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().Align(lipgloss.Center).Render(strings.Join(lines, "\n")),
	)
	return normalBorder.Width(w).Height(h).
		BorderForeground(colDimGray).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Foreground(colDimGray).Render("Dashboard"),
			body,
		))
}
