package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) ensureSearchChatInventory() tea.Cmd {
	if m.searchChatInventoryLoaded || m.searchChatInventoryLoading {
		return nil
	}
	m.searchChatInventoryLoading = true
	return loadSearchChatInventoryCmd(m.clientID, m.app.CurrentUserName, m.userID)
}

func (m Model) openNewChatPicker() (Model, tea.Cmd) {
	m.app.PendingForwardText = ""
	m.app.NewChatMode = true
	m.app.NewChatComposePending = false
	m.app.NewChatSelectedUsers = nil
	m.app.NewChatLocalResults = nil
	m.app.NewChatDirectoryQuery = ""
	m.app.UserSearchPopupMode = true
	m.app.UserSearchMode = true
	m.app.UserSearchQuery = ""
	m.app.UserSearchStatus = ""
	m.app.UserSearchLocalResults = nil
	m.app.UserSearchMemberResults = nil
	m.app.UserSearchMessageResults = nil
	m.app.UserSearchChannelResults = nil
	m.app.UserSearchDirectoryResults = nil
	m.app.UserSearchSelectedIndex = 0
	m.app.UserSearchLoading = false
	m.userSearchInput.Placeholder = "Find people by name or email..."
	m.userSearchInput.SetValue("")
	m.userSearchInput.Focus()
	m.updateNewChatLocalResults()
	return m, tea.Batch(textinput.Blink, m.ensureSearchChatInventory())
}

func newChatUserKey(user User) string {
	if email := newChatUserEmail(user); email != "" {
		return strings.ToLower(email)
	}
	if identity := userDirectoryIdentity(user); identity != "" {
		return strings.ToLower(identity)
	}
	return strings.ToLower(strings.TrimSpace(user.DisplayName))
}

func newChatUserEmail(user User) string {
	if user.UserPrincipalName != nil && strings.TrimSpace(*user.UserPrincipalName) != "" {
		return strings.TrimSpace(*user.UserPrincipalName)
	}
	if user.Mail != nil {
		return strings.TrimSpace(*user.Mail)
	}
	return ""
}

func newChatUserLabel(user User) string {
	if strings.TrimSpace(user.DisplayName) != "" {
		return strings.TrimSpace(user.DisplayName)
	}
	if email := newChatUserEmail(user); email != "" {
		return email
	}
	return strings.TrimSpace(user.ID)
}

func memberAsDirectoryUser(member ChatMember) (User, bool) {
	user := User{}
	if member.DisplayName != nil {
		user.DisplayName = strings.TrimSpace(*member.DisplayName)
	}
	if member.UserID != nil {
		user.ID = strings.TrimSpace(*member.UserID)
	}
	if member.Email != nil && strings.TrimSpace(*member.Email) != "" {
		email := strings.TrimSpace(*member.Email)
		user.UserPrincipalName = &email
	}
	return user, userDirectoryIdentity(user) != ""
}

func (m Model) knownUsersForNewChat() []User {
	byKey := make(map[string]User)
	for _, chat := range m.knownChatsForSearch() {
		for _, member := range chat.Members {
			user, ok := memberAsDirectoryUser(member)
			if !ok || (m.userID != "" && strings.EqualFold(user.ID, m.userID)) {
				continue
			}
			key := newChatUserKey(user)
			if key != "" {
				byKey[key] = user
			}
		}
	}
	users := make([]User, 0, len(byKey))
	for _, user := range byKey {
		users = append(users, user)
	}
	sort.SliceStable(users, func(i, j int) bool {
		return strings.ToLower(newChatUserLabel(users[i])) < strings.ToLower(newChatUserLabel(users[j]))
	})
	return users
}

func newChatUserTarget(user User) searchTarget {
	return searchTarget{Text: []string{user.DisplayName, newChatUserEmail(user), user.ID}}
}

func (m *Model) updateNewChatLocalResults() {
	query := parseSearchQuery(m.app.UserSearchQuery)
	type scoredUser struct {
		user  User
		score int
	}
	var matches []scoredUser
	for _, user := range m.knownUsersForNewChat() {
		score := 0
		matched := true
		if len(query.Terms) > 0 {
			score, matched = query.Match(newChatUserTarget(user))
		}
		if matched {
			matches = append(matches, scoredUser{user: user, score: score})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return strings.ToLower(newChatUserLabel(matches[i].user)) < strings.ToLower(newChatUserLabel(matches[j].user))
	})
	m.app.NewChatLocalResults = nil
	for index, match := range matches {
		if index >= 40 {
			break
		}
		m.app.NewChatLocalResults = append(m.app.NewChatLocalResults, match.user)
	}
}

func (m Model) newChatUserSelected(user User) bool {
	key := newChatUserKey(user)
	for _, selected := range m.app.NewChatSelectedUsers {
		if newChatUserKey(selected) == key {
			return true
		}
	}
	return false
}

func (m Model) toggleNewChatUser(user User) Model {
	key := newChatUserKey(user)
	if key == "" {
		return m
	}
	selected := make([]User, 0, len(m.app.NewChatSelectedUsers)+1)
	removed := false
	for _, current := range m.app.NewChatSelectedUsers {
		if newChatUserKey(current) == key {
			removed = true
			continue
		}
		selected = append(selected, current)
	}
	if !removed {
		selected = append(selected, user)
	}
	m.app.NewChatSelectedUsers = selected
	m.app.UserSearchQuery = ""
	m.app.UserSearchDirectoryResults = nil
	m.app.NewChatDirectoryQuery = ""
	m.userSearchInput.SetValue("")
	m.updateNewChatLocalResults()
	m.app.UserSearchSelectedIndex = 0
	return m
}

func exactDirectoryUser(query string) (User, bool) {
	query = strings.TrimSpace(query)
	if strings.ContainsAny(query, " \t\r\n") || !strings.Contains(query, "@") {
		return User{}, false
	}
	return User{ID: query, DisplayName: query, UserPrincipalName: &query}, true
}

func (m Model) startNewChatCreation() (Model, tea.Cmd) {
	if m.app.UserSearchLoading && m.app.NewChatComposePending {
		return m, nil
	}
	if len(m.app.NewChatSelectedUsers) == 0 {
		m.app.UserSearchStatus = "Select at least one participant first."
		return m, nil
	}
	m.app.NewChatComposePending = true
	m.app.UserSearchLoading = true
	m.app.UserSearchStatus = "Creating chat..."
	return m, createParticipantsChatCmd(m.clientID, m.userID, m.app.NewChatSelectedUsers)
}

func (m Model) closeNewChatPicker() Model {
	m.app.UserSearchPopupMode = false
	m.app.UserSearchMode = false
	m.app.UserSearchLoading = false
	m.app.NewChatMode = false
	m.app.NewChatComposePending = false
	m.app.NewChatSelectedUsers = nil
	m.app.NewChatLocalResults = nil
	m.app.UserSearchDirectoryResults = nil
	m.app.UserSearchStatus = ""
	m.userSearchInput.Blur()
	return m
}

func (m Model) handleNewChatInputModeKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+j":
		return m.startNewChatCreation()
	case "esc":
		return m.closeNewChatPicker(), nil
	case "down", "up", "tab":
		m.app.UserSearchMode = false
		m.userSearchInput.Blur()
		m.app.UserSearchSelectedIndex = 0
		return m, nil
	case "enter":
		m.app.UserSearchQuery = strings.TrimSpace(m.userSearchInput.Value())
		m.updateNewChatLocalResults()
		items := m.getUserSearchItems()
		if len(items) == 0 || items[0].DirUser == nil {
			m.app.UserSearchStatus = "No participant selected; try an exact email address."
			return m, nil
		}
		return m.toggleNewChatUser(*items[0].DirUser), nil
	}
	return m, nil
}

func (m Model) handleNewChatNavigationKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	items := m.getUserSearchItems()
	switch msg.String() {
	case "esc", "q":
		return m.closeNewChatPicker(), nil
	case "ctrl+j":
		return m.startNewChatCreation()
	case "j", "down", "tab":
		if len(items) > 0 && m.app.UserSearchSelectedIndex < len(items)-1 {
			m.app.UserSearchSelectedIndex++
		}
		return m, nil
	case "k", "up", "shift+tab":
		if m.app.UserSearchSelectedIndex > 0 {
			m.app.UserSearchSelectedIndex--
		}
		return m, nil
	case "/":
		m.app.UserSearchMode = true
		m.userSearchInput.Focus()
		return m, textinput.Blink
	case " ", "space", "enter":
		if m.app.UserSearchSelectedIndex >= 0 && m.app.UserSearchSelectedIndex < len(items) && items[m.app.UserSearchSelectedIndex].DirUser != nil {
			m = m.toggleNewChatUser(*items[m.app.UserSearchSelectedIndex].DirUser)
			m.app.UserSearchMode = true
			m.userSearchInput.Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m Model) renderNewChatPopup(w, h int) string {
	title := lipgloss.NewStyle().Foreground(colCyan).Bold(true).Render("New Chat")
	instructions := lipgloss.NewStyle().Foreground(colDimGray).Render(
		"Type a name/email · Enter add/remove · arrows then Space toggle · Ctrl+Enter create",
	)

	selectedLabels := make([]string, 0, len(m.app.NewChatSelectedUsers))
	for _, user := range m.app.NewChatSelectedUsers {
		selectedLabels = append(selectedLabels, newChatUserLabel(user))
	}
	selectedText := "Selected: none"
	if len(selectedLabels) > 0 {
		selectedText = fmt.Sprintf("Selected (%d): %s", len(selectedLabels), strings.Join(selectedLabels, ", "))
	}
	selectedText = truncate(selectedText, max(10, w-8))

	var list strings.Builder
	list.WriteString(title + "\n")
	list.WriteString(instructions + "\n")
	list.WriteString(lipgloss.NewStyle().Foreground(colGreen).Bold(len(selectedLabels) > 0).Render(selectedText) + "\n\n")

	items := m.getUserSearchItems()
	listHeight := h - 12
	if listHeight < 2 {
		listHeight = 2
	}
	if m.app.UserSearchSelectedIndex >= len(items) {
		m.app.UserSearchSelectedIndex = len(items) - 1
	}
	if m.app.UserSearchSelectedIndex < 0 {
		m.app.UserSearchSelectedIndex = 0
	}
	start := 0
	if m.app.UserSearchSelectedIndex >= listHeight {
		start = m.app.UserSearchSelectedIndex - listHeight + 1
	}
	end := min(len(items), start+listHeight)
	if len(items) == 0 {
		list.WriteString(lipgloss.NewStyle().Foreground(colDimGray).Render("  Type to find a person, or enter an exact email address.") + "\n")
	} else {
		for index := start; index < end; index++ {
			item := items[index]
			if item.DirUser == nil {
				continue
			}
			checked := "[ ]"
			if m.newChatUserSelected(*item.DirUser) {
				checked = "[x]"
			}
			label := newChatUserLabel(*item.DirUser)
			email := newChatUserEmail(*item.DirUser)
			label, _ = bidiVisualLine(label)
			if email != "" && !strings.EqualFold(label, email) {
				label += " · " + email
			}
			tagText := "[" + item.Source + "]"
			available := max(8, w-8-lipgloss.Width(tagText)-lipgloss.Width(checked)-5)
			label = truncate(label, available)
			tag := lipgloss.NewStyle().Foreground(colDimGray).Render(tagText)
			prefix := "  "
			lineStyle := lipgloss.NewStyle()
			if index == m.app.UserSearchSelectedIndex {
				prefix = "> "
				lineStyle = lineStyle.Background(colDarkGray).Foreground(colWhite).Bold(true)
			}
			line := fmt.Sprintf("%s%s %s %s", prefix, checked, label, tag)
			list.WriteString(lineStyle.Render(line) + "\n")
		}
	}
	for lines := max(1, end-start); lines < listHeight; lines++ {
		list.WriteString("\n")
	}

	m.userSearchInput.Width = max(10, w-12)
	borderCol := colCyan
	if !m.app.UserSearchMode {
		borderCol = colDimGray
	}
	inputBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderCol).
		Width(w - 6).Height(3).
		Render(lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Foreground(borderCol).Bold(true).Render("+ "),
			m.userSearchInput.View(),
		))

	status := m.app.UserSearchStatus
	if status == "" && m.searchChatInventoryLoading {
		status = "Loading older chats and known participants..."
	}
	if status != "" {
		list.WriteString(lipgloss.NewStyle().Foreground(colYellow).Italic(true).Render("  "+status) + "\n")
	} else {
		list.WriteString("\n")
	}
	list.WriteString(inputBox)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colCyan).
		Padding(1, 2).
		Width(w).Height(h).
		Render(list.String())
}
