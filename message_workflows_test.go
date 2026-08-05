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

func TestNormalMailStyleMessageBindings(t *testing.T) {
	t.Run("i marks read", func(t *testing.T) {
		model := newWorkflowTestModel()
		model, cmd := model.handleNormalModeKey(filterTestKey('i'))
		if cmd == nil || model.app.Status != "Marking chat read..." {
			t.Fatalf("i returned cmd=%v status=%q, want explicit read action", cmd != nil, model.app.Status)
		}
	})

	for _, key := range []rune{'c', 'C'} {
		t.Run(string(key)+" composes", func(t *testing.T) {
			model := newWorkflowTestModel()
			model, _ = model.handleNormalModeKey(filterTestKey(key))
			if !model.app.InputMode || model.app.ReplyToMessage != nil {
				t.Fatalf("%c did not open a normal composer", key)
			}
		})
	}

	for _, key := range []rune{'r', 'R'} {
		t.Run(string(key)+" replies", func(t *testing.T) {
			model := newWorkflowTestModel()
			model, _ = model.handleNormalModeKey(filterTestKey(key))
			if !model.app.InputMode || model.app.ReplyToMessage == nil || model.app.ReplyToMessage.ID != "newest" {
				t.Fatalf("%c did not reply to newest loaded message", key)
			}
		})
	}

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
	for _, key := range []rune{'r', 'R'} {
		model := newWorkflowTestModel()
		model.app.MessageSelectionMode = true
		model.app.MessageSelectedIndex = 1
		model, _ = model.handleMessageSelectionModeKey(filterTestKey(key))
		if model.app.ReplyToMessage == nil || model.app.ReplyToMessage.ID != "older" || !model.app.InputMode {
			t.Fatalf("%c did not reply to the selected message", key)
		}
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

	model := newWorkflowTestModel()
	model.app.MessageSelectionMode = true
	model, _ = model.handleMessageSelectionModeKey(filterTestKey('+'))
	if !model.app.ReactionMode {
		t.Fatal("+ did not preserve reaction access")
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
		threadActionReply:    "r",
		threadActionForward:  "f",
		threadActionRead:     "i",
		threadActionUnread:   "u",
		threadActionFavorite: "*",
		threadActionCapture:  "a",
	}
	for action, key := range want {
		if keys[action] != key {
			t.Fatalf("action %s key = %q, want %q", action, keys[action], key)
		}
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
