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
	threadActionAnalyze     threadActionID = "analyze"
	threadActionCopyLink    threadActionID = "copy-link"
	threadActionArtifacts   threadActionID = "artifacts"
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
		{Key: "R", Label: "Reply to latest message", ID: threadActionReply},
		{Key: "f", Label: "Forward latest message", ID: threadActionForward},
		{Key: "r", Label: "Mark read", ID: threadActionRead},
		{Key: "u", Label: "Mark unread", ID: threadActionUnread},
		{Key: "*", Label: "Toggle favorite", ID: threadActionFavorite},
		{Key: "a", Label: "Capture in configured thread list", ID: threadActionCapture},
		{Key: "e", Label: "Export complete Markdown transcript", ID: threadActionExport},
		{Key: "A", Label: "Analyze complete thread with configured command", ID: threadActionAnalyze},
		{Key: "y", Label: "Copy Teams link", ID: threadActionCopyLink},
		{Key: "t", Label: "Choose recording or transcript", ID: threadActionArtifacts},
	}
}

func threadActionBinding(action threadActionID) string {
	switch action {
	case threadActionOpenBrowser:
		return keyThreadOpenBrowser
	case threadActionOpenTeams:
		return keyThreadOpenApp
	case threadActionCompose:
		return keyThreadCompose
	case threadActionReply:
		return keyThreadReply
	case threadActionForward:
		return keyThreadForward
	case threadActionRead:
		return keyThreadRead
	case threadActionUnread:
		return keyThreadUnread
	case threadActionFavorite:
		return keyThreadFavorite
	case threadActionCapture:
		return keyThreadCapture
	case threadActionExport:
		return keyThreadExport
	case threadActionAnalyze:
		return keyThreadAnalyze
	case threadActionCopyLink:
		return keyThreadCopyLink
	case threadActionArtifacts:
		return keyThreadArtifacts
	default:
		return ""
	}
}

func (m Model) configuredThreadActions() []threadAction {
	actions := threadActions()
	for index := range actions {
		actions[index].Key = m.keybindings.Primary(threadActionBinding(actions[index].ID))
	}
	return actions
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

func threadActionAdvancesChat(action threadActionID) bool {
	switch action {
	case threadActionOpenBrowser,
		threadActionOpenTeams,
		threadActionRead,
		threadActionUnread,
		threadActionFavorite,
		threadActionCapture,
		threadActionExport,
		threadActionAnalyze,
		threadActionCopyLink:
		return true
	default:
		return false
	}
}

// nextVisibleChatID captures the post-action target before read/favorite state
// can rebuild the filtered sidebar and change row indexes.
func (m Model) nextVisibleChatID(chatID string) string {
	if chatID == "" || len(m.app.Chats) < 2 {
		return ""
	}
	for index := range m.app.Chats {
		if m.app.Chats[index].ID == chatID {
			return m.app.Chats[(index+1)%len(m.app.Chats)].ID
		}
	}
	return ""
}

func (m Model) advanceAfterThreadAction(nextChatID string, actionCmd tea.Cmd) (Model, tea.Cmd) {
	if nextChatID == "" || !m.app.SetSelectedChatID(nextChatID) {
		return m, actionCmd
	}

	m.channelSelectedIndex = -1
	m.app.SelectedChannelTeamID = ""
	m.app.SelectedChannelID = ""
	m.app.SearchMode = false
	m.app.SearchActive = false
	m.app.SearchQuery = ""
	m.app.SnapToBottom = true
	m.pendingChatGoto = false
	delete(m.manuallyUnread, nextChatID)
	m = m.markRead()

	var loadCmd tea.Cmd
	m, loadCmd = m.loadChatMessages(nextChatID)
	if actionCmd == nil {
		return m, loadCmd
	}
	if loadCmd == nil {
		return m, actionCmd
	}
	return m, tea.Batch(actionCmd, loadCmd)
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
	nextChatID := ""
	if threadActionAdvancesChat(action) {
		nextChatID = m.nextVisibleChatID(chatValue.ID)
	}
	advance := func(updated Model, cmd tea.Cmd) (Model, tea.Cmd) {
		return updated.advanceAfterThreadAction(nextChatID, cmd)
	}

	switch action {
	case threadActionOpenBrowser:
		if chatURL == "" {
			m.app.SetStatus("This chat did not include a Teams URL", 4*time.Second)
			return m, nil
		}
		m.app.SetStatus("Opening chat in browser...", 0)
		return advance(m, openURLCmd(teamsWebURL(chatURL), m.app.BrowserCommand, m.app.YoutrackCommand, m.app.GitlabCommand))

	case threadActionOpenTeams:
		deepLink := teamsDesktopURL(chatURL)
		if deepLink == "" {
			m.app.SetStatus("This chat did not include a Teams desktop link", 4*time.Second)
			return m, nil
		}
		m.app.SetStatus("Opening chat in Teams...", 0)
		return advance(m, openWithCommandCmd(deepLink, m.app.TeamsAppCommand))

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
		return advance(m, setChatReadStateCmd(m.clientID, chatValue.ID, m.userID, false))

	case threadActionUnread:
		m.app.SetStatus("Marking chat unread...", 0)
		return advance(m, setChatReadStateCmd(m.clientID, chatValue.ID, m.userID, true))

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
		if nextChatID != "" {
			return advance(m, nil)
		}
		return m.reconcileSelectedChatConversation()

	case threadActionCapture:
		m.app.SetStatus("Capturing thread...", 0)
		path := m.app.ThreadCaptureFile
		if normalizeThreadCaptureFormat(m.app.ThreadCaptureFormat) == ThreadCaptureOrg {
			path = m.app.ThreadCaptureOrgFile
		}
		return advance(m, captureChatCmd(chatValue, m.app.ThreadCaptureFormat, path))

	case threadActionExport:
		m.app.SetStatus("Exporting complete chat history...", 0)
		return advance(m, exportChatMarkdownCmd(m.clientID, chatValue, m.app.ExportDirectory))

	case threadActionAnalyze:
		m.app.SetStatus("Exporting complete chat history for "+m.app.ThreadAnalysisAgent+" analysis...", 0)
		return advance(m, analyzeChatThreadCmd(
			m.clientID,
			chatValue,
			m.app.ExportDirectory,
			m.app.ThreadAnalysisAgent,
			m.app.ThreadAnalysisCommand,
		))

	case threadActionCopyLink:
		if chatURL == "" {
			m.app.SetStatus("This chat did not include a Teams URL", 4*time.Second)
			return m, nil
		}
		if err := clipboard.WriteAll(chatURL); err != nil {
			m.app.SetStatus("Could not copy Teams link: "+err.Error(), 4*time.Second)
		} else {
			m.app.SetStatus("Teams link copied", 3*time.Second)
			return advance(m, nil)
		}

	case threadActionArtifacts:
		return m.openConversationArtifacts()
	}
	return m, nil
}

func (m Model) handleThreadActionPopupKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	actions := threadActions()
	key := m.keyName(keyContextThreadActions, msg)
	switch key {
	case "esc", "q":
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

	switch key {
	case "C":
		key = "c"
	case "F":
		key = "f"
	case "i":
		key = "r"
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
	for index, action := range m.configuredThreadActions() {
		cursor := "  "
		style := lipgloss.NewStyle()
		if index == m.app.ThreadActionSelectedIndex {
			cursor = "› "
			style = style.Foreground(colCyan).Bold(true)
		}
		lines = append(lines, style.Render(fmt.Sprintf("%s%s  %s", cursor, action.Key, action.Label)))
	}
	lines = append(lines, "", lipgloss.NewStyle().Foreground(colDimGray).Render(
		m.keybindings.Display(keyListSelect)+" run · "+m.keybindings.Display(keyListClose)+" cancel"))
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colGreen).
		Padding(1, 2).
		Width(w - 2).
		MaxHeight(h).
		Render(strings.Join(lines, "\n"))
}
