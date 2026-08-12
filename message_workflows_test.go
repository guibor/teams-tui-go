package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func workflowTestMessage(id, sender, body string) Message {
	content := "<p>" + body + "</p>"
	senderID := "sender-" + id
	return Message{
		ID:              id,
		CreatedDateTime: "2026-08-05T12:34:00Z",
		From: &MessageFrom{User: &MessageUser{
			ID:          &senderID,
			DisplayName: &sender,
		}},
		Body: &MessageBody{Content: &content},
	}
}

func newWorkflowTestModel() Model {
	app := NewApp()
	app.Chats = []Chat{{ID: "chat-1"}}
	app.SelectedIndex = 0
	app.Messages = []Message{
		workflowTestMessage("newest", "Alice", "Newest message"),
		workflowTestMessage("older", "Bob", "Older message"),
	}
	model := NewModel(app, "client", "user")
	model.latestChats = append([]Chat(nil), app.Chats...)
	model.stableChatOrder = []string{"chat-1"}
	return model
}

func newWorkflowChatListModel(ids ...string) Model {
	model := newWorkflowTestModel()
	chats := make([]Chat, 0, len(ids))
	for index, id := range ids {
		name := "Chat " + id
		preview := workflowTestMessage("message-"+id, "Sender", "Message "+id)
		preview.ChatID = id
		chats = append(chats, Chat{
			ID:                 id,
			CachedDisplayName:  &name,
			LastMessagePreview: &preview,
		})
		model.lastMsgID[id] = preview.ID
		if index == 0 {
			for messageIndex := range model.app.Messages {
				model.app.Messages[messageIndex].ChatID = id
			}
		}
	}
	model.stableChatOrder = append(model.stableChatOrder[:0], ids...)
	model.app.Chats = append([]Chat(nil), chats...)
	model.latestChats = append([]Chat(nil), chats...)
	model.chatCache = make(map[string]Chat, len(chats))
	for _, chat := range chats {
		model.chatCache[chat.ID] = chat
	}
	if len(ids) > 0 {
		model.app.SetSelectedChatID(ids[0])
		model.app.MessagesConversationID = ids[0]
	}
	return model
}

func TestNormalMailStyleMessageBindings(t *testing.T) {
	for _, key := range []rune{'r', 'i'} {
		t.Run(string(key)+" marks read", func(t *testing.T) {
			model := newWorkflowTestModel()
			model, cmd := model.handleNormalModeKey(filterTestKey(key))
			if cmd == nil || model.app.Status != "Marking chat read..." || model.app.InputMode {
				t.Fatalf("%c returned cmd=%v status=%q input=%v, want explicit read action", key, cmd != nil, model.app.Status, model.app.InputMode)
			}
		})
	}

	for _, key := range []rune{'c', 'C'} {
		t.Run(string(key)+" composes", func(t *testing.T) {
			model := newWorkflowTestModel()
			model, _ = model.handleNormalModeKey(filterTestKey(key))
			if !model.app.InputMode || model.app.ReplyToMessage != nil {
				t.Fatalf("%c did not open a normal composer", key)
			}
		})
	}

	t.Run("R replies", func(t *testing.T) {
		model := newWorkflowTestModel()
		model, _ = model.handleNormalModeKey(filterTestKey('R'))
		if !model.app.InputMode || model.app.ReplyToMessage == nil || model.app.ReplyToMessage.ID != "newest" {
			t.Fatal("R did not reply to newest loaded message")
		}
	})

	for _, key := range []rune{'f', 'F'} {
		t.Run(string(key)+" forwards", func(t *testing.T) {
			model := newWorkflowTestModel()
			model, _ = model.handleNormalModeKey(filterTestKey(key))
			if !model.app.UserSearchPopupMode || !strings.Contains(model.app.PendingForwardText, "Newest message") {
				t.Fatalf("%c did not open the forward destination chooser", key)
			}
		})
	}
}

func TestReadBindingsAdvanceToNextVisibleChat(t *testing.T) {
	for _, key := range []rune{'r', 'i'} {
		t.Run(string(key), func(t *testing.T) {
			model := newWorkflowChatListModel("chat-1", "chat-2", "chat-3")
			model, cmd := model.handleNormalModeKey(filterTestKey(key))
			selected := model.app.GetSelectedChat()
			if cmd == nil || selected == nil || selected.ID != "chat-2" {
				t.Fatalf("%c selected %#v with cmd=%v, want chat-2 and read command", key, selected, cmd != nil)
			}
			if model.app.MessagesConversationID != "chat-2" {
				t.Fatalf("%c transcript owner = %q, want chat-2", key, model.app.MessagesConversationID)
			}
		})
	}
}

func TestReadCompletionKeepsOverlaySelectionAndTranscriptInSync(t *testing.T) {
	model := newWorkflowChatListModel("chat-1", "chat-2", "chat-3")
	model.app.UnreadOverlay = true
	model = model.rebuildChatList()
	model, cmd := model.handleNormalModeKey(filterTestKey('r'))
	if cmd == nil || model.app.SelectedChatID != "chat-2" || model.app.MessagesConversationID != "chat-2" {
		t.Fatalf("read action did not advance coherently: selected=%q transcript=%q", model.app.SelectedChatID, model.app.MessagesConversationID)
	}
	model, _ = model.updateInternal(MsgChatReadStateChanged{ChatID: "chat-1"})
	if model.app.SelectedChatID != "chat-2" || model.app.MessagesConversationID != "chat-2" {
		t.Fatalf("read completion mixed identities: selected=%q transcript=%q", model.app.SelectedChatID, model.app.MessagesConversationID)
	}
	for _, chat := range model.app.Chats {
		if chat.ID == "chat-1" {
			t.Fatal("read chat remained visible under unread overlay")
		}
	}
}

func TestReadAdvanceDoesNotSkipAfterUnreadFilterRemovesChat(t *testing.T) {
	model := newWorkflowChatListModel("chat-1", "chat-2", "chat-3")
	filter := newChatListFilter()
	filter.ReadState = ChatReadUnread
	model.app.ActiveChatFilter = filter
	model = model.rebuildChatList()
	model.app.SetSelectedChatID("chat-1")

	model, _ = model.handleNormalModeKey(filterTestKey('r'))
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != "chat-2" {
		t.Fatalf("initial read advance selected %#v, want chat-2", selected)
	}

	model, _ = model.updateInternal(MsgChatReadStateChanged{ChatID: "chat-1"})
	if len(model.app.Chats) != 2 || model.app.Chats[0].ID != "chat-2" || model.app.Chats[1].ID != "chat-3" {
		t.Fatalf("filtered chats after read = %#v, want chat-2 then chat-3", model.app.Chats)
	}
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != "chat-2" {
		t.Fatalf("post-filter selection = %#v, want chat-2 without skipping", selected)
	}
	if model.app.MessagesConversationID != "chat-2" {
		t.Fatalf("post-filter transcript owner = %q, want chat-2", model.app.MessagesConversationID)
	}
}

func TestReadActionWrapsToFirstVisibleChat(t *testing.T) {
	model := newWorkflowChatListModel("chat-1", "chat-2", "chat-3")
	model.app.SetSelectedChatID("chat-3")
	model.app.MessagesConversationID = "chat-3"
	model, _ = model.handleNormalModeKey(filterTestKey('r'))
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != "chat-1" {
		t.Fatalf("read at end selected %#v, want wrapped chat-1", selected)
	}
	if model.app.MessagesConversationID != "chat-1" {
		t.Fatalf("wrapped transcript owner = %q, want chat-1", model.app.MessagesConversationID)
	}
}

func TestThreadActionAdvancePolicy(t *testing.T) {
	for _, action := range []threadActionID{
		threadActionOpenBrowser,
		threadActionOpenTeams,
		threadActionRead,
		threadActionUnread,
		threadActionFavorite,
		threadActionCapture,
		threadActionExport,
		threadActionAnalyze,
		threadActionCopyLink,
	} {
		if !threadActionAdvancesChat(action) {
			t.Errorf("action %s should advance", action)
		}
	}
	for _, action := range []threadActionID{
		threadActionCompose,
		threadActionReply,
		threadActionForward,
		threadActionArtifacts,
	} {
		if threadActionAdvancesChat(action) {
			t.Errorf("interactive action %s should retain the current chat", action)
		}
	}
}

func TestCaptureActionAdvancesWhileComposeRetainsChat(t *testing.T) {
	model := newWorkflowChatListModel("chat-1", "chat-2")
	model.app.ThreadCaptureFile = t.TempDir() + "/threads.md"
	model, cmd := model.executeThreadAction(threadActionCapture)
	if cmd == nil {
		t.Fatal("capture action returned no command")
	}
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != "chat-2" {
		t.Fatalf("capture selected %#v, want chat-2", selected)
	}

	model = newWorkflowChatListModel("chat-1", "chat-2")
	model, _ = model.executeThreadAction(threadActionCompose)
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != "chat-1" {
		t.Fatalf("compose selected %#v, want current chat-1", selected)
	}
	if !model.app.InputMode {
		t.Fatal("compose did not enter input mode")
	}
}

func TestArtifactChooserAdvancesAfterOpeningResource(t *testing.T) {
	model := newWorkflowChatListModel("chat-1", "chat-2")
	model.app.ArtifactPopupMode = true
	model.app.Artifacts = []ConversationArtifact{{
		Kind: ConversationRecording,
		URL:  "https://teams.microsoft.com/resource",
	}}

	model, cmd := model.handleConversationArtifactPopupKey(filterTestKey('o'))
	if cmd == nil {
		t.Fatal("artifact open returned no command")
	}
	if model.app.ArtifactPopupMode {
		t.Fatal("artifact chooser remained open after dispatch")
	}
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != "chat-2" {
		t.Fatalf("artifact open selected %#v, want chat-2", selected)
	}
	if model.app.MessagesConversationID != "chat-2" {
		t.Fatalf("artifact open transcript owner = %q, want chat-2", model.app.MessagesConversationID)
	}
}

func TestDisplacedChatBindingsRemainAvailable(t *testing.T) {
	model := newWorkflowTestModel()
	model, _ = model.handleNormalModeKey(filterTestKey('v'))
	if !model.app.ChatFilterPopupMode {
		t.Fatal("v did not open the advanced chat filter")
	}

	model = newWorkflowTestModel()
	model, _ = model.handleNormalModeKey(filterTestKey('*'))
	if !model.favourites["chat-1"] {
		t.Fatal("* did not toggle the selected chat favorite")
	}
}

func TestMessageSelectionMailStyleBindings(t *testing.T) {
	model := newWorkflowTestModel()
	model.app.MessageSelectionMode = true
	model.app.MessageSelectedIndex = 1
	model, _ = model.handleMessageSelectionModeKey(filterTestKey('R'))
	if model.app.ReplyToMessage == nil || model.app.ReplyToMessage.ID != "older" || !model.app.InputMode {
		t.Fatal("R did not reply to the selected message")
	}

	model = newWorkflowTestModel()
	model.app.MessageSelectionMode = true
	model, cmd := model.handleMessageSelectionModeKey(filterTestKey('r'))
	if cmd == nil || model.app.InputMode || model.app.Status != "Marking chat read..." {
		t.Fatalf("r did not mark the selected message's conversation read: cmd=%v input=%v status=%q", cmd != nil, model.app.InputMode, model.app.Status)
	}

	for _, key := range []rune{'f', 'F'} {
		model := newWorkflowTestModel()
		model.app.MessageSelectionMode = true
		model.app.MessageSelectedIndex = 1
		model, _ = model.handleMessageSelectionModeKey(filterTestKey(key))
		if !model.app.UserSearchPopupMode || !strings.Contains(model.app.PendingForwardText, "Older message") {
			t.Fatalf("%c did not forward the selected message", key)
		}
	}

	model = newWorkflowTestModel()
	model.app.MessageSelectionMode = true
	model, _ = model.handleMessageSelectionModeKey(filterTestKey('+'))
	if !model.app.ReactionMode {
		t.Fatal("+ did not preserve reaction access")
	}
}

func TestMessagePopupDistinguishesReadFromReply(t *testing.T) {
	model := newWorkflowTestModel()
	model.app.MessagePopupMode = true
	model.app.MessageSelectedIndex = 1
	model, _ = model.handleMessagePopupKey(filterTestKey('R'))
	if model.app.ReplyToMessage == nil || model.app.ReplyToMessage.ID != "older" || !model.app.InputMode {
		t.Fatal("R did not reply to the popup message")
	}

	model = newWorkflowTestModel()
	model.app.MessagePopupMode = true
	model.app.MessageSelectedIndex = 1
	model, cmd := model.handleMessagePopupKey(filterTestKey('r'))
	if cmd == nil || model.app.InputMode || model.app.Status != "Marking chat read..." {
		t.Fatalf("r did not mark the popup conversation read: cmd=%v input=%v status=%q", cmd != nil, model.app.InputMode, model.app.Status)
	}
}

func TestForwardDestinationOpensEditableComposer(t *testing.T) {
	model := newWorkflowTestModel()
	target := Chat{ID: "chat-2"}
	model.app.Chats = append(model.app.Chats, target)
	model.latestChats = append(model.latestChats, target)
	model.stableChatOrder = append(model.stableChatOrder, target.ID)
	model.app.UserSearchPopupMode = true
	model.app.UserSearchLocalResults = []Chat{target}
	model.app.PendingForwardText = "forward body"

	model, _ = model.handleUserSearchNavigationKey(tea.KeyMsg{Type: tea.KeyEnter})
	selected := model.app.GetSelectedChat()
	if selected == nil || selected.ID != target.ID {
		t.Fatalf("forward selected %#v, want target chat", selected)
	}
	if !model.app.InputMode || model.textarea.Value() != "forward body" {
		t.Fatalf("forward composer input=%v body=%q", model.app.InputMode, model.textarea.Value())
	}
	if model.app.PendingForwardText != "" || model.app.UserSearchPopupMode {
		t.Fatal("forward chooser state was not cleared after selecting a destination")
	}
}

func TestForwardChooserCancelDiscardsPendingCopy(t *testing.T) {
	model := newWorkflowTestModel()
	model.app.UserSearchPopupMode = true
	model.app.UserSearchMode = false
	model.app.PendingForwardText = "forward body"

	model, _ = model.handleUserSearchNavigationKey(tea.KeyMsg{Type: tea.KeyEsc})
	if model.app.PendingForwardText != "" || model.app.UserSearchPopupMode {
		t.Fatal("cancel did not discard the pending forward")
	}
}

func TestForwardChooserPreloadsAndAcceptsComponentSearchLocalChat(t *testing.T) {
	model := newWorkflowChatListModel("source", "target", "other")
	targetName := "Quarterly Planning"
	otherName := "General Updates"
	model.app.Chats[1].CachedDisplayName = &targetName
	model.app.Chats[2].CachedDisplayName = &otherName
	model.latestChats = append([]Chat(nil), model.app.Chats...)
	for _, chat := range model.app.Chats {
		model.chatCache[chat.ID] = chat
	}

	model, _ = model.openChatChooser("forward body")
	if len(model.app.UserSearchLocalResults) != 3 {
		t.Fatalf("empty forward query returned %d local chats, want 3", len(model.app.UserSearchLocalResults))
	}

	model.userSearchInput.SetValue("q p")
	model.app.UserSearchQuery = "q p"
	model.updateUserSearchLocalResults()
	if len(model.app.UserSearchLocalResults) != 1 || model.app.UserSearchLocalResults[0].ID != "target" {
		t.Fatalf("component-search destination results = %#v, want target", model.app.UserSearchLocalResults)
	}

	model, cmd := model.handleUserSearchInputModeKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("accepting component-search destination returned no load/compose command")
	}
	if selected := model.app.GetSelectedChat(); selected == nil || selected.ID != "target" {
		t.Fatalf("accepted component-search destination selected %#v, want target", selected)
	}
	if !model.app.InputMode || model.textarea.Value() != "forward body" {
		t.Fatalf("accepted destination input=%v body=%q", model.app.InputMode, model.textarea.Value())
	}
}

func TestCreatedDirectChatContinuesPendingForward(t *testing.T) {
	model := newWorkflowTestModel()
	model.app.PendingForwardText = "forward to a new direct chat"
	model.app.UserSearchPopupMode = true
	model.app.UserSearchMode = false
	target := Chat{ID: "new-direct", ChatType: "oneOnOne"}

	model, _ = model.updateInternal(MsgCreateChatDone{Chat: &target})
	selected := model.app.GetSelectedChat()
	if selected == nil || selected.ID != target.ID {
		t.Fatalf("created forward selected %#v, want new direct chat", selected)
	}
	if !model.app.InputMode || model.textarea.Value() != "forward to a new direct chat" {
		t.Fatalf("created forward composer input=%v body=%q", model.app.InputMode, model.textarea.Value())
	}
	if model.app.PendingForwardText != "" || model.app.UserSearchPopupMode {
		t.Fatal("created forward did not clear chooser state")
	}
}

func TestThreadActionKeysMatchMailStyleBindings(t *testing.T) {
	keys := make(map[threadActionID]string)
	for _, action := range threadActions() {
		keys[action.ID] = action.Key
	}
	want := map[threadActionID]string{
		threadActionCompose:  "c",
		threadActionReply:    "R",
		threadActionForward:  "f",
		threadActionRead:     "r",
		threadActionUnread:   "u",
		threadActionFavorite: "*",
		threadActionCapture:  "a",
		threadActionAnalyze:  "A",
	}
	for action, key := range want {
		if keys[action] != key {
			t.Fatalf("action %s key = %q, want %q", action, keys[action], key)
		}
	}
}

func TestThreadActionPopupLowercaseRMarksRead(t *testing.T) {
	model := newWorkflowTestModel()
	model.app.ThreadActionPopupMode = true
	model, cmd := model.handleThreadActionPopupKey(filterTestKey('r'))
	if cmd == nil || model.app.InputMode || model.app.Status != "Marking chat read..." {
		t.Fatalf("r returned cmd=%v input=%v status=%q, want read action", cmd != nil, model.app.InputMode, model.app.Status)
	}
}

func TestThreadActionPopupAcceptsUppercaseMessageActions(t *testing.T) {
	tests := []struct {
		key    rune
		assert func(*testing.T, Model)
	}{
		{'C', func(t *testing.T, model Model) {
			if !model.app.InputMode || model.app.ReplyToMessage != nil {
				t.Fatal("C did not open a normal composer")
			}
		}},
		{'R', func(t *testing.T, model Model) {
			if !model.app.InputMode || model.app.ReplyToMessage == nil {
				t.Fatal("R did not open a reply composer")
			}
		}},
		{'F', func(t *testing.T, model Model) {
			if !model.app.UserSearchPopupMode || model.app.PendingForwardText == "" {
				t.Fatal("F did not open the forward destination chooser")
			}
		}},
	}
	for _, test := range tests {
		t.Run(string(test.key), func(t *testing.T) {
			model := newWorkflowTestModel()
			model.app.ThreadActionPopupMode = true
			model, _ = model.handleThreadActionPopupKey(filterTestKey(test.key))
			if model.app.ThreadActionPopupMode {
				t.Fatalf("%c did not close the action popup", test.key)
			}
			test.assert(t, model)
		})
	}
}

func TestForwardedMessageMarkdownIncludesReadableMetadataAndAttachments(t *testing.T) {
	message := workflowTestMessage("source", "Alice", "Forward this")
	name := "notes.txt"
	url := "https://example.test/notes.txt"
	message.Attachments = []MessageAttachment{{Name: &name, ContentURL: &url}}

	forwarded := forwardedMessageMarkdown(message)
	for _, want := range []string{"**Forwarded message**", "**From:** Alice", "**Date:**", "Forward this", "[notes.txt](https://example.test/notes.txt)"} {
		if !strings.Contains(forwarded, want) {
			t.Fatalf("forwarded Markdown missing %q:\n%s", want, forwarded)
		}
	}
	if strings.Contains(forwarded, "<p>") {
		t.Fatalf("forwarded Markdown retained HTML: %s", forwarded)
	}
}
