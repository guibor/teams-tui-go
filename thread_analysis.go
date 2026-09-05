package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var threadAnalysisDestinations = []string{"terminal", "emacs", "codex-app"}

func analysisDisplayName(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "terminal":
		return "Ghostty terminal"
	case "emacs":
		return "Emacs agent-shell"
	case "codex-app":
		return "Codex app"
	case "default":
		return "Default model"
	default:
		return value
	}
}

func (m Model) threadAnalysisSummary() string {
	return analysisDisplayName(m.app.ThreadAnalysisDestination) + " · " +
		analysisDisplayName(m.app.ThreadAnalysisAgent) + " · " +
		analysisDisplayName(m.app.ThreadAnalysisModel)
}

func (m Model) launchThreadAnalysis(chat Chat, destination, model string) (Model, tea.Cmd) {
	m.app.ThreadAnalysisPopupMode = false
	m.app.SetStatus("Exporting complete chat for "+analysisDisplayName(model)+" in "+analysisDisplayName(destination)+"...", 0)
	return m, analyzeChatThreadCmd(m.clientID, chat, m.app.ExportDirectory,
		m.app.ThreadAnalysisAgent, destination, model, m.app.ThreadAnalysisCommand)
}

func (m Model) threadAnalysisChoices() []string {
	if m.app.ThreadAnalysisStage == 0 {
		return threadAnalysisDestinations
	}
	if len(m.app.ThreadAnalysisModels) == 0 {
		return []string{"default"}
	}
	return m.app.ThreadAnalysisModels
}

func (m Model) handleThreadAnalysisPopupKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	choices := m.threadAnalysisChoices()
	switch msg.String() {
	case "esc", "q":
		m.app.ThreadAnalysisPopupMode = false
		return m, nil
	case "j", "down", "tab":
		m.app.ThreadAnalysisSelectedIndex = (m.app.ThreadAnalysisSelectedIndex + 1) % len(choices)
		return m, nil
	case "k", "up", "shift+tab":
		m.app.ThreadAnalysisSelectedIndex = (m.app.ThreadAnalysisSelectedIndex - 1 + len(choices)) % len(choices)
		return m, nil
	case "enter":
		choice := choices[m.app.ThreadAnalysisSelectedIndex]
		if m.app.ThreadAnalysisStage == 0 {
			m.app.ThreadAnalysisPendingDestination = choice
			m.app.ThreadAnalysisStage = 1
			m.app.ThreadAnalysisSelectedIndex = 0
			for index, candidate := range m.app.ThreadAnalysisModels {
				if candidate == m.app.ThreadAnalysisModel {
					m.app.ThreadAnalysisSelectedIndex = index
					break
				}
			}
			return m, nil
		}
		chat := m.app.GetSelectedChat()
		if chat == nil {
			m.app.ThreadAnalysisPopupMode = false
			m.app.SetStatus("Select a chat first", 3*time.Second)
			return m, nil
		}
		return m.launchThreadAnalysis(*chat, m.app.ThreadAnalysisPendingDestination, choice)
	}
	return m, nil
}

func (m Model) renderThreadAnalysisPopup(w, h int) string {
	title := "Choose analysis destination"
	if m.app.ThreadAnalysisStage == 1 {
		title = "Choose analysis model"
	}
	lines := []string{
		lipgloss.NewStyle().Foreground(colYellow).Bold(true).Render(title),
		lipgloss.NewStyle().Foreground(colDimGray).Render("Configured: " + m.threadAnalysisSummary()),
		lipgloss.NewStyle().Foreground(colDimGray).Render("Command: " + m.app.ThreadAnalysisCommand), "",
	}
	for index, choice := range m.threadAnalysisChoices() {
		cursor, style := "  ", lipgloss.NewStyle()
		if index == m.app.ThreadAnalysisSelectedIndex {
			cursor = "› "
			style = style.Foreground(colCyan).Bold(true)
		}
		lines = append(lines, style.Render(cursor+analysisDisplayName(choice)))
	}
	lines = append(lines, "", lipgloss.NewStyle().Foreground(colDimGray).Render("enter select · esc cancel"))
	return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(colGreen).
		Padding(1, 2).Width(w - 2).MaxHeight(h).Render(strings.Join(lines, "\n"))
}
