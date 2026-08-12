package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type ConversationArtifactKind string

const (
	ConversationRecording  ConversationArtifactKind = "recording"
	ConversationTranscript ConversationArtifactKind = "transcript"
)

type ConversationArtifact struct {
	Kind       ConversationArtifactKind
	MessageID  string
	Title      string
	URL        string
	CreatedAt  time.Time
	DirectLink bool
}

func collectConversationArtifacts(messages []Message, conversationURL string) []ConversationArtifact {
	artifacts := make([]ConversationArtifact, 0)
	for _, message := range messages {
		if message.EventDetail == nil {
			continue
		}
		kind := message.EventDetail.shortType()
		artifact := ConversationArtifact{
			MessageID: message.ID,
			CreatedAt: searchTime(message.CreatedDateTime),
		}
		switch kind {
		case "callRecording":
			artifact.Kind = ConversationRecording
			artifact.Title = strings.TrimSpace(message.EventDetail.CallRecordingDisplayName)
			if artifact.Title == "" {
				artifact.Title = message.SystemEventSummary()
			}
			artifact.URL = strings.TrimSpace(message.EventDetail.CallRecordingURL)
			artifact.DirectLink = artifact.URL != ""
		case "callTranscript":
			artifact.Kind = ConversationTranscript
			artifact.Title = message.SystemEventSummary()
		default:
			continue
		}
		if artifact.URL == "" {
			artifact.URL = strings.TrimSpace(message.WebURL)
		}
		if artifact.URL == "" {
			artifact.URL = strings.TrimSpace(conversationURL)
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts
}

func conversationArtifactCounts(messages []Message) (recordings, transcripts int) {
	for _, artifact := range collectConversationArtifacts(messages, "") {
		switch artifact.Kind {
		case ConversationRecording:
			recordings++
		case ConversationTranscript:
			transcripts++
		}
	}
	return
}

func (m Model) openConversationArtifacts() (Model, tea.Cmd) {
	if m.channelSelectedIndex >= 0 {
		m.app.SetStatus("Meeting resources currently support chats", 4*time.Second)
		return m, nil
	}
	chat := m.app.GetSelectedChat()
	if chat == nil {
		m.app.SetStatus("Select a chat first", 3*time.Second)
		return m, nil
	}
	messages := m.knownMessagesForSearch(*chat)
	m.app.Artifacts = collectConversationArtifacts(messages, teamsChatURL(chat))
	if len(m.app.Artifacts) == 0 {
		m.app.SetStatus("No recordings or transcripts in loaded conversation history", 4*time.Second)
		return m, nil
	}
	m.app.ArtifactSelectedIndex = 0
	m.app.ArtifactPopupMode = true
	return m, nil
}

func (m Model) selectedConversationArtifact() *ConversationArtifact {
	index := m.app.ArtifactSelectedIndex
	if index < 0 || index >= len(m.app.Artifacts) {
		return nil
	}
	return &m.app.Artifacts[index]
}

func (m Model) openSelectedConversationArtifact() (Model, tea.Cmd) {
	artifact := m.selectedConversationArtifact()
	if artifact == nil || artifact.URL == "" {
		m.app.SetStatus("This resource did not include an openable link", 4*time.Second)
		return m, nil
	}
	nextChatID := ""
	if chat := m.app.GetSelectedChat(); chat != nil {
		nextChatID = m.nextVisibleChatID(chat.ID)
	}
	m.app.ArtifactPopupMode = false
	m.app.SetStatus("Opening "+string(artifact.Kind)+"...", 0)
	return m.advanceAfterThreadAction(
		nextChatID,
		openURLCmd(teamsWebURL(artifact.URL), m.app.BrowserCommand, m.app.YoutrackCommand, m.app.GitlabCommand),
	)
}

func (m Model) handleConversationArtifactPopupKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch m.keyName(keyContextArtifacts, msg) {
	case "esc", "q":
		m.app.ArtifactPopupMode = false
		return m, nil
	case "j", "down", "tab":
		if len(m.app.Artifacts) > 0 {
			m.app.ArtifactSelectedIndex = (m.app.ArtifactSelectedIndex + 1) % len(m.app.Artifacts)
		}
		return m, nil
	case "k", "up", "shift+tab":
		if len(m.app.Artifacts) > 0 {
			m.app.ArtifactSelectedIndex--
			if m.app.ArtifactSelectedIndex < 0 {
				m.app.ArtifactSelectedIndex = len(m.app.Artifacts) - 1
			}
		}
		return m, nil
	case "enter", "o":
		return m.openSelectedConversationArtifact()
	case "y":
		artifact := m.selectedConversationArtifact()
		if artifact == nil || artifact.URL == "" {
			return m, nil
		}
		if err := clipboard.WriteAll(artifact.URL); err != nil {
			m.app.SetStatus("Could not copy resource link: "+err.Error(), 4*time.Second)
		} else {
			m.app.SetStatus("Resource link copied", 3*time.Second)
			nextChatID := ""
			if chat := m.app.GetSelectedChat(); chat != nil {
				nextChatID = m.nextVisibleChatID(chat.ID)
			}
			m.app.ArtifactPopupMode = false
			return m.advanceAfterThreadAction(nextChatID, nil)
		}
		return m, nil
	}
	return m, nil
}

func (m Model) renderConversationArtifactPopup(w, h int) string {
	if w < 54 {
		w = 54
	}
	lines := []string{
		lipgloss.NewStyle().Foreground(colYellow).Bold(true).Render(fmt.Sprintf("Recordings and transcripts (%d)", len(m.app.Artifacts))),
		lipgloss.NewStyle().Foreground(colDimGray).Render(m.activeConversationTitle()),
		"",
	}
	visibleRows := h - 8
	if visibleRows < 1 {
		visibleRows = 1
	}
	start := m.app.ArtifactSelectedIndex - visibleRows/2
	if start < 0 {
		start = 0
	}
	if maxStart := len(m.app.Artifacts) - visibleRows; start > maxStart && maxStart >= 0 {
		start = maxStart
	}
	end := start + visibleRows
	if end > len(m.app.Artifacts) {
		end = len(m.app.Artifacts)
	}
	if start > 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(colDimGray).Render(fmt.Sprintf("  ↑ %d newer", start)))
	}
	for index := start; index < end; index++ {
		artifact := m.app.Artifacts[index]
		cursor := "  "
		style := lipgloss.NewStyle()
		if index == m.app.ArtifactSelectedIndex {
			cursor = "› "
			style = style.Foreground(colCyan).Bold(true)
		}
		kind := "TRANSCRIPT"
		if artifact.Kind == ConversationRecording {
			kind = "RECORDING"
		}
		when := ""
		if !artifact.CreatedAt.IsZero() {
			when = artifact.CreatedAt.Format("2006-01-02 15:04") + " · "
		}
		source := "Teams event"
		if artifact.DirectLink {
			source = "direct link"
		}
		line := fmt.Sprintf("%s%s · %s%s · %s", cursor, kind, when, artifact.Title, source)
		lines = append(lines, style.Render(ansi.Truncate(line, w-6, "…")))
	}
	if end < len(m.app.Artifacts) {
		lines = append(lines, lipgloss.NewStyle().Foreground(colDimGray).Render(fmt.Sprintf("  ↓ %d older", len(m.app.Artifacts)-end)))
	}
	lines = append(lines, "", lipgloss.NewStyle().Foreground(colDimGray).Render(
		m.keybindings.Display(keyArtifactOpen)+" open · "+m.keybindings.Display(keyArtifactCopy)+" copy link · "+m.keybindings.Display(keyListClose)+" close"))
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colGreen).
		Padding(1, 2).
		Width(w - 2).
		MaxHeight(h).
		Render(strings.Join(lines, "\n"))
}
