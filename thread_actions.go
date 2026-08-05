package main

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type threadActionID string

const (
	threadActionOpenBrowser threadActionID = "open-browser"
	threadActionOpenTeams   threadActionID = "open-teams"
	threadActionCompose     threadActionID = "compose"
	threadActionReply       threadActionID = "reply"
	threadActionForward     threadActionID = "forward"
	threadActionRead        threadActionID = "read"
	threadActionUnread      threadActionID = "unread"
	threadActionFavorite    threadActionID = "favorite"
	threadActionCapture     threadActionID = "capture"
	threadActionExport      threadActionID = "export"
	threadActionCopyLink    threadActionID = "copy-link"
)

type threadAction struct {
	Key   string
	Label string
	ID    threadActionID
}

func threadActions() []threadAction {
	return []threadAction{
		{Key: "o", Label: "Open in browser", ID: threadActionOpenBrowser},
		{Key: "O", Label: "Open in Teams desktop", ID: threadActionOpenTeams},
		{Key: "c", Label: "Compose message", ID: threadActionCompose},
		{Key: "r", Label: "Reply to latest message", ID: threadActionReply},
		{Key: "f", Label: "Forward latest message", ID: threadActionForward},
		{Key: "i", Label: "Mark read", ID: threadActionRead},
		{Key: "u", Label: "Mark unread", ID: threadActionUnread},
		{Key: "*", Label: "Toggle favorite", ID: threadActionFavorite},
		{Key: "w", Label: "Capture in Markdown thread list", ID: threadActionCapture},
		{Key: "e", Label: "Export complete Markdown transcript", ID: threadActionExport},
		{Key: "y", Label: "Copy Teams link", ID: threadActionCopyLink},
	}
}

func teamsDesktopURL(webURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(webURL))
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || !isTeamsWebHost(parsed.Hostname()) {
		return ""
	}
	parsed.Scheme = "msteams"
	return parsed.String()
}

func isTeamsWebHost(host string) bool {
	return strings.EqualFold(host, "teams.microsoft.com") ||
		strings.EqualFold(host, "teams.cloud.microsoft")
}

// teamsWebURL bypasses the /l launcher endpoint, which can dispatch the
// msteams:// protocol, and routes the same opaque target through Teams web.
func teamsWebURL(webURL string) string {
	raw := strings.TrimSpace(webURL)
	parsed, err := url.Parse(raw)
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || !isTeamsWebHost(parsed.Hostname()) {
		return raw
	}
	if strings.HasPrefix(parsed.Fragment, "/l/") || !strings.HasPrefix(parsed.EscapedPath(), "/l/") {
		return raw
	}

	origin := &url.URL{Scheme: parsed.Scheme, Host: parsed.Host, Path: "/"}
	route := parsed.EscapedPath()
	if parsed.RawQuery != "" {
		route += "?" + parsed.RawQuery
	}
	return origin.String() + "#" + route
}

func (m Model) executeThreadAction(action threadActionID) (Model, tea.Cmd) {
	if m.channelSelectedIndex >= 0 {
		m.app.SetStatus("Thread actions currently support chats", 4*time.Second)
		return m, nil
	}
	chat := m.app.GetSelectedChat()
	if chat == nil {
		m.app.SetStatus("Select a chat first", 3*time.Second)
		return m, nil
	}
	chatValue := *chat
	chatURL := teamsChatURL(chat)

	switch action {
	case threadActionOpenBrowser:
		if chatURL == "" {
			m.app.SetStatus("This chat did not include a Teams URL", 4*time.Second)
			return m, nil
		}
		m.app.SetStatus("Opening chat in browser...", 0)
		return m, openURLCmd(teamsWebURL(chatURL), m.app.BrowserCommand, m.app.YoutrackCommand, m.app.GitlabCommand)

	case threadActionOpenTeams:
		deepLink := teamsDesktopURL(chatURL)
		if deepLink == "" {
			m.app.SetStatus("This chat did not include a Teams desktop link", 4*time.Second)
			return m, nil
		}
		m.app.SetStatus("Opening chat in Teams...", 0)
		return m, openWithCommandCmd(deepLink, m.app.TeamsAppCommand)

	case threadActionCompose:
		return m.beginCompose("")

	case threadActionReply:
		message, ok := m.newestLoadedMessage()
		if !ok {
			m.app.SetStatus("No loaded message to reply to", 3*time.Second)
			return m, nil
		}
		return m.beginReply(message)

	case threadActionForward:
		message, ok := m.newestLoadedMessage()
		if !ok {
			m.app.SetStatus("No loaded message to forward", 3*time.Second)
			return m, nil
		}
		return m.beginForward(message)

	case threadActionRead:
		m.app.SetStatus("Marking chat read...", 0)
		return m, setChatReadStateCmd(m.clientID, chatValue.ID, m.userID, false)

	case threadActionUnread:
		m.app.SetStatus("Marking chat unread...", 0)
		return m, setChatReadStateCmd(m.clientID, chatValue.ID, m.userID, true)

	case threadActionFavorite:
		if m.favourites[chatValue.ID] {
			delete(m.favourites, chatValue.ID)
			m.app.SetStatus("Removed from favorites: "+chatExportTitle(chatValue), 3*time.Second)
		} else {
			m.favourites[chatValue.ID] = true
			m.app.SetStatus("Added to favorites: "+chatExportTitle(chatValue), 3*time.Second)
		}
		_ = SaveFavourites(m.favourites)
		m = m.rebuildChatList()
		selected := m.app.GetSelectedChat()
		if selected == nil {
			m.app.ClearMessagesConversation()
			return m, nil
		}
		if selected.ID != chatValue.ID || !m.app.MessagesBelongTo(selected.ID) {
			m.app.SnapToBottom = true
			return m.loadChatMessages(selected.ID)
		}
		return m, nil

	case threadActionCapture:
		m.app.SetStatus("Capturing thread...", 0)
		return m, captureChatMarkdownCmd(chatValue, m.app.ThreadCaptureFile)

	case threadActionExport:
		m.app.SetStatus("Exporting complete chat history...", 0)
		return m, exportChatMarkdownCmd(m.clientID, chatValue, m.app.ExportDirectory)

	case threadActionCopyLink:
		if chatURL == "" {
			m.app.SetStatus("This chat did not include a Teams URL", 4*time.Second)
			return m, nil
		}
		if err := clipboard.WriteAll(chatURL); err != nil {
			m.app.SetStatus("Could not copy Teams link: "+err.Error(), 4*time.Second)
		} else {
			m.app.SetStatus("Teams link copied", 3*time.Second)
		}
	}
	return m, nil
}

func (m Model) handleThreadActionPopupKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	actions := threadActions()
	key := msg.String()
	switch key {
	case "esc", "q", "a":
		m.app.ThreadActionPopupMode = false
		return m, nil
	case "j", "down", "tab":
		m.app.ThreadActionSelectedIndex = (m.app.ThreadActionSelectedIndex + 1) % len(actions)
		return m, nil
	case "k", "up", "shift+tab":
		m.app.ThreadActionSelectedIndex--
		if m.app.ThreadActionSelectedIndex < 0 {
			m.app.ThreadActionSelectedIndex = len(actions) - 1
		}
		return m, nil
	case "enter":
		index := m.app.ThreadActionSelectedIndex
		if index < 0 || index >= len(actions) {
			return m, nil
		}
		m.app.ThreadActionPopupMode = false
		return m.executeThreadAction(actions[index].ID)
	}

	if key == "C" || key == "R" || key == "F" {
		key = strings.ToLower(key)
	}
	for _, action := range actions {
		if key == action.Key {
			m.app.ThreadActionPopupMode = false
			return m.executeThreadAction(action.ID)
		}
	}
	return m, nil
}

func (m Model) renderThreadActionPopup(w, h int) string {
	if w < 48 {
		w = 48
	}
	chatName := "Selected chat"
	if chat := m.app.GetSelectedChat(); chat != nil {
		chatName = chatExportTitle(*chat)
	}
	lines := []string{
		lipgloss.NewStyle().Foreground(colYellow).Bold(true).Render("Thread actions"),
		lipgloss.NewStyle().Foreground(colDimGray).Render(chatName),
		"",
	}
	for index, action := range threadActions() {
		cursor := "  "
		style := lipgloss.NewStyle()
		if index == m.app.ThreadActionSelectedIndex {
			cursor = "› "
			style = style.Foreground(colCyan).Bold(true)
		}
		lines = append(lines, style.Render(fmt.Sprintf("%s%s  %s", cursor, action.Key, action.Label)))
	}
	lines = append(lines, "", lipgloss.NewStyle().Foreground(colDimGray).Render("Enter run · Esc cancel"))
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colGreen).
		Padding(1, 2).
		Width(w - 2).
		MaxHeight(h).
		Render(strings.Join(lines, "\n"))
}
