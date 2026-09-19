package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTenantEndpointsFollowConfiguration(t *testing.T) {
	for _, tenant := range []string{"", "example.onmicrosoft.com"} {
		t.Setenv("TENANT_ID", tenant)
		want := tenant
		if want == "" {
			want = "common"
		}
		prefix := "https://login.microsoftonline.com/" + want + "/oauth2/v2.0/"
		if deviceCodeURL() != prefix+"devicecode" || tokenURL() != prefix+"token" {
			t.Fatalf("wrong endpoints for tenant %q", tenant)
		}
	}
}

func TestSnoozeExpiryInvalidatesCachedView(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := newWorkflowChatListModel("first", "second")
	m.width, m.height = 100, 30
	m.snoozed["second"] = time.Now().Add(time.Hour)
	m = m.rebuildChatList()
	_ = m.View()
	m.snoozed["second"] = time.Now().Add(-time.Second)
	updated, _ := m.Update(MsgTick{})
	m = updated.(Model)
	if !m.viewCache.dirty || len(m.app.Chats) != 2 {
		t.Fatal("expired snooze must restore the hidden chat and repaint")
	}
	if m.app.SelectedChatID != "first" || !m.app.MessagesBelongTo("first") {
		t.Fatal("waking another chat changed the active conversation")
	}
}

func TestVisualBellExpiryInvalidatesCachedView(t *testing.T) {
	m := newWorkflowChatListModel("first")
	m.width, m.height = 100, 30
	m.app.TriggerVisualBell()
	_ = m.View()
	past := time.Now().Add(-time.Second)
	m.app.VisualBellUntil = &past
	updated, _ := m.Update(MsgTick{})
	m = updated.(Model)
	if !m.viewCache.dirty || m.app.VisualBellUntil != nil {
		t.Fatal("expired visual bell remained cached")
	}
}

func TestIdleCacheStillRepairsMismatchedConversation(t *testing.T) {
	m := newWorkflowChatListModel("first", "second")
	m.width, m.height = 100, 30
	_ = m.View()
	m.app.MessagesConversationID = "second"
	m.app.Messages = []Message{{ID: "wrong", ChatID: "second"}}
	updated, cmd := m.Update(MsgTick{})
	m = updated.(Model)
	if cmd == nil || !m.viewCache.dirty || !m.app.MessagesBelongTo("first") {
		t.Fatal("cached idle frame bypassed transcript ownership reconciliation")
	}
}

func TestForwardedRenderingKeepsSystemDetailsAndHiddenChatNames(t *testing.T) {
	m := newWorkflowChatListModel("first")
	name := "Filtered project"
	m.chatCache["hidden"] = Chat{ID: "hidden", CachedDisplayName: &name}
	content := `{"originalMessageContent":"<p>Project update</p>","originalConversationId":"hidden","originalMessageSender":{"user":{"displayName":"Alex"}}}`
	kind := "forwardedMessageReference"
	body := `<attachment id="reference"></attachment>`
	msg := Message{Body: &MessageBody{Content: &body}, Attachments: []MessageAttachment{{ID: "reference", ContentType: &kind, Content: &content}}}
	if text := stripANSI(m.messagePlainText(&msg)); !strings.Contains(text, name) || !strings.Contains(text, "Project update") {
		t.Fatalf("forwarded preview lost hidden source-chat metadata: %q", text)
	}
	if len(viewableAttachments(msg)) != 0 {
		t.Fatal("forwarded quote became a downloadable attachment")
	}
	event := Message{MessageType: "systemEventMessage", EventDetail: &EventMessageDetail{
		ODataType: "#microsoft.graph.callEndedEventMessageDetail", CallEventType: "meeting", CallDuration: "PT10S",
	}}
	if got := m.messagePlainText(&event); got != "Meeting ended (10s)" {
		t.Fatalf("system event regressed to a placeholder: %q", got)
	}
}

func TestDownloadOnlyShortcutIsConfigurableAndModeScoped(t *testing.T) {
	m := newWorkflowChatListModel("first")
	m.keybindings, _ = NewKeyMap(KeyBindingConfig{keyMessageViewDownload: KeyList{"x"}})
	url, name := "https://example.com/report.pdf", "report.pdf"
	m.app.Messages = []Message{{ID: "msg", ChatID: "first", Attachments: []MessageAttachment{{ContentURL: &url, Name: &name}}}}
	m.app.MessageSelectedIndex = 0
	m.app.MessagePopupMode = true
	m.app.AttachmentCursorMode = true
	m.app.Features.FilePreview = true
	_, cmd := m.handleMessagePopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd != nil {
		t.Fatal("overridden download key still dispatched")
	}
	m, cmd = m.handleMessagePopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if cmd == nil || !strings.HasPrefix(m.app.MessagePopupStatus, "Saving:") || m.app.DeleteConfirmMode {
		t.Fatal("custom download-only key did not start saving in the popup")
	}
	if m.keybindings.Canonical(keyContextMessageSelect, "d") != "d" {
		t.Fatal("download binding changed the message-delete shortcut")
	}
}

func TestStreamedDownloadPublishesOnlyCompleteFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/partial" {
			w.Header().Set("Content-Length", "100")
		}
		fmt.Fprint(w, "payload")
	}))
	defer server.Close()
	if downloadClient.Timeout <= graphHTTPClient.Timeout {
		t.Fatal("attachment transfers still use the short API timeout")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "download")
	if err := DownloadFile("unused", server.URL, path); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{path, filepath.Join(dir, "new-download")} {
		if err := DownloadFile("unused", server.URL+"/partial", target); err == nil {
			t.Fatal("incomplete transfer reported success")
		}
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "payload" {
		t.Fatal("failed transfer replaced the complete file")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatal("failed transfer left a partial file or temporary artifact")
	}
}
