package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type snoozeChoice struct {
	Key   string
	Label string
	Until func(Model, time.Time) time.Time
}

func parseLocalClock(value string, fallbackHour int) (int, int) {
	parsed, err := time.Parse("15:04", strings.TrimSpace(value))
	if err != nil {
		return fallbackHour, 0
	}
	return parsed.Hour(), parsed.Minute()
}

func nextLocalClock(now time.Time, clock string, fallbackHour int, forceTomorrow bool) time.Time {
	hour, minute := parseLocalClock(clock, fallbackHour)
	day := now
	if forceTomorrow {
		day = day.AddDate(0, 0, 1)
	}
	target := time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, now.Location())
	if !forceTomorrow && !target.After(now) {
		startHour, startMinute := parseLocalClock("07:00", 7)
		tomorrow := now.AddDate(0, 0, 1)
		target = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), startHour, startMinute, 0, 0, now.Location())
	}
	return target
}

func snoozeChoices() []snoozeChoice {
	return []snoozeChoice{
		{Key: "m", Label: "10 minutes", Until: func(_ Model, now time.Time) time.Time { return now.Add(10 * time.Minute) }},
		{Key: "1", Label: "1 hour", Until: func(_ Model, now time.Time) time.Time { return now.Add(time.Hour) }},
		{Key: "3", Label: "3 hours", Until: func(_ Model, now time.Time) time.Time { return now.Add(3 * time.Hour) }},
		{Key: "e", Label: "End of workday", Until: func(m Model, now time.Time) time.Time {
			target := nextLocalClock(now, m.app.WorkdayEnd, 18, false)
			endHour, endMinute := parseLocalClock(m.app.WorkdayEnd, 18)
			endToday := time.Date(now.Year(), now.Month(), now.Day(), endHour, endMinute, 0, 0, now.Location())
			if !endToday.After(now) {
				target = nextLocalClock(now, m.app.WorkdayStart, 7, true)
			}
			return target
		}},
		{Key: "t", Label: "Tomorrow morning", Until: func(m Model, now time.Time) time.Time { return nextLocalClock(now, m.app.WorkdayStart, 7, true) }},
		{Key: "w", Label: "Next week", Until: func(m Model, now time.Time) time.Time {
			return nextLocalClock(now.AddDate(0, 0, 6), m.app.WorkdayStart, 7, true)
		}},
		{Key: "u", Label: "Unsnooze", Until: nil},
	}
}

func (m Model) chatSnoozed(chatID string, now time.Time) bool {
	until, ok := m.snoozed[chatID]
	return ok && until.After(now)
}

func (m *Model) wakeChat(chatID string) bool {
	if _, ok := m.snoozed[chatID]; !ok {
		return false
	}
	delete(m.snoozed, chatID)
	_ = SaveSnoozedChats(m.snoozed)
	return true
}

func (m *Model) pruneExpiredSnoozes(now time.Time) bool {
	changed := false
	for id, until := range m.snoozed {
		if !until.After(now) {
			delete(m.snoozed, id)
			changed = true
		}
	}
	if changed {
		_ = SaveSnoozedChats(m.snoozed)
	}
	return changed
}

func (m Model) applySnooze(until time.Time) (Model, tea.Cmd) {
	chat := m.app.GetSelectedChat()
	if chat == nil {
		return m, nil
	}
	chatID := chat.ID
	nextID := m.nextVisibleChatID(chatID)
	if until.IsZero() {
		delete(m.snoozed, chatID)
		m.app.SetStatus("Unsnoozed "+chatExportTitle(*chat), 3*time.Second)
	} else {
		m.snoozed[chatID] = until
		m.app.SetStatus("Snoozed until "+until.Format("Mon 15:04"), 4*time.Second)
	}
	_ = SaveSnoozedChats(m.snoozed)
	m.app.SnoozePopupMode = false
	m = m.rebuildChatList()
	if m.app.SetSelectedChatID(chatID) {
		return m.reconcileSelectedChatConversation()
	}
	if nextID != "" {
		m.app.SetSelectedChatID(nextID)
	}
	return m.reconcileSelectedChatConversation()
}

func (m Model) quickSnooze() (Model, tea.Cmd) {
	minutes := m.app.DefaultSnoozeMinutes
	if minutes <= 0 {
		minutes = 180
	}
	return m.applySnooze(time.Now().Add(time.Duration(minutes) * time.Minute))
}

func (m Model) handleSnoozePopupKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	choices := snoozeChoices()
	key := m.keyName(keyContextSnooze, msg)
	if key == "esc" {
		m.app.SnoozePopupMode = false
		return m, nil
	}
	if key == "j" || key == "down" {
		m.app.SnoozeSelectedIndex = (m.app.SnoozeSelectedIndex + 1) % len(choices)
		return m, nil
	}
	if key == "k" || key == "up" {
		m.app.SnoozeSelectedIndex = (m.app.SnoozeSelectedIndex - 1 + len(choices)) % len(choices)
		return m, nil
	}
	if key == "enter" {
		key = choices[m.app.SnoozeSelectedIndex].Key
	}
	for _, choice := range choices {
		if key == choice.Key {
			if choice.Until == nil {
				return m.applySnooze(time.Time{})
			}
			return m.applySnooze(choice.Until(m, time.Now()))
		}
	}
	return m, nil
}

func (m Model) renderSnoozePopup(w, h int) string {
	choices := snoozeChoices()
	lines := []string{lipgloss.NewStyle().Bold(true).Render("Snooze chat"), ""}
	for index, choice := range choices {
		prefix := "  "
		style := lipgloss.NewStyle()
		if index == m.app.SnoozeSelectedIndex {
			prefix = "> "
			style = style.Bold(true).Foreground(colCyan)
		}
		lines = append(lines, style.Render(fmt.Sprintf("%s%s  %s", prefix, choice.Key, choice.Label)))
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colYellow).
		Padding(1, 2).Width(w).MaxHeight(h).Render(strings.Join(lines, "\n"))
}
