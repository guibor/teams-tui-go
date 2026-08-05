package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) beginCompose(initialText string) (Model, tea.Cmd) {
	m.app.InputMode = true
	m.app.InputBuffer = initialText
	m.app.EditingMessageID = nil
	m.app.ReplyToMessage = nil
	m.app.ChannelReplyToID = ""
	m.app.ComposedImages = nil
	m.app.ComposedFiles = nil
	m.app.MessageSelectionMode = false
	m.app.MessagePopupMode = false
	m.app.AttachmentCursorMode = false
	m.app.PendingForwardText = ""
	m.textarea.Reset()
	m.textarea.SetValue(initialText)
	return m, m.textarea.Focus()
}

func (m Model) beginReply(message Message) (Model, tea.Cmd) {
	m, cmd := m.beginCompose("")
	if m.activeChannelEntry() != nil {
		rootID := message.ID
		if message.IsReply && message.ReplyToID != "" {
			rootID = message.ReplyToID
		}
		m.app.ChannelReplyToID = rootID
	} else {
		reference := message
		m.app.ReplyToMessage = &reference
	}
	return m, cmd
}

func (m Model) beginForward(message Message) (Model, tea.Cmd) {
	m.app.MessageSelectionMode = false
	m.app.MessagePopupMode = false
	return m.openChatChooser(forwardedMessageMarkdown(message))
}

func (m Model) openChatChooser(forwardText string) (Model, tea.Cmd) {
	m.app.PendingForwardText = forwardText
	m.app.UserSearchPopupMode = true
	m.app.UserSearchMode = true
	m.app.UserSearchQuery = ""
	m.app.UserSearchStatus = ""
	m.app.UserSearchLocalResults = nil
	m.app.UserSearchChannelResults = nil
	m.app.UserSearchDirectoryResults = nil
	m.app.UserSearchSelectedIndex = 0
	m.app.UserSearchLoading = false
	m.userSearchInput.SetValue("")
	m.userSearchInput.Focus()
	return m, textinput.Blink
}

func (m Model) newestLoadedMessage() (Message, bool) {
	if len(m.app.Messages) == 0 {
		return Message{}, false
	}
	return m.app.Messages[0], true
}

func (m Model) markActiveConversationRead() (Model, tea.Cmd) {
	if m.channelSelectedIndex >= 0 {
		if entry := m.activeChannelEntry(); entry != nil {
			latestID := m.lastMsgID[entry.channelID]
			if latestID == "" && len(m.app.Messages) > 0 {
				latestID = m.app.Messages[0].ID
			}
			if latestID != "" {
				m.lastReadMsgID[entry.channelID] = latestID
				delete(m.manuallyUnread, entry.channelID)
				m.app.SetStatus("Marked channel read locally", 3*time.Second)
			}
		}
		return m, nil
	}
	if m.app.GetSelectedChat() != nil {
		return m.executeThreadAction(threadActionRead)
	}
	return m, nil
}

func forwardedMessageMarkdown(message Message) string {
	header := fmt.Sprintf("**From:** %s  \n**Date:** %s", markdownSender(message), markdownMessageTime(message.CreatedDateTime))
	sections := []string{"**Forwarded message**", header}
	if subject := strings.TrimSpace(message.Subject); subject != "" {
		sections = append(sections, "**Subject:** "+subject)
	}
	sections = append(sections, markdownMessageBody(&message))
	if attachments := strings.TrimSpace(markdownAttachments(message)); attachments != "" {
		sections = append(sections, attachments)
	}
	return strings.Join(sections, "\n\n")
}
