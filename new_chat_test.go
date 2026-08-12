package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func testDirectoryUser(id, name, email string) User {
	return User{ID: id, DisplayName: name, UserPrincipalName: &email}
}

func TestNormalUpperNOpensParticipantPicker(t *testing.T) {
	model := newWorkflowTestModel()
	model, cmd := model.handleNormalModeKey(filterTestKey('N'))
	if !model.app.UserSearchPopupMode || !model.app.NewChatMode || !model.app.UserSearchMode {
		t.Fatalf("N did not open participant picker: popup=%v new=%v input=%v", model.app.UserSearchPopupMode, model.app.NewChatMode, model.app.UserSearchMode)
	}
	if cmd == nil {
		t.Fatal("N did not return input/inventory commands")
	}
}

func TestNewChatPickerFindsKnownUserAndStartsComposerCreation(t *testing.T) {
	model := newWorkflowTestModel()
	name := "Alice Example"
	email := "alice@example.com"
	userID := "alice-id"
	chatName := "Alice Example"
	chat := Chat{ID: "alice-chat", CachedDisplayName: &chatName, Members: []ChatMember{{UserID: &userID, DisplayName: &name, Email: &email}}}
	model.searchChatInventory = []Chat{chat}
	model.searchChatInventoryLoaded = true
	model, _ = model.openNewChatPicker()
	model.userSearchInput.SetValue("alice ex")
	model.app.UserSearchQuery = "alice ex"
	model.updateNewChatLocalResults()

	model, _ = model.handleNewChatInputModeKey(tea.KeyMsg{Type: tea.KeyEnter})
	if len(model.app.NewChatSelectedUsers) != 1 || newChatUserEmail(model.app.NewChatSelectedUsers[0]) != email {
		t.Fatalf("known participant was not selected: %#v", model.app.NewChatSelectedUsers)
	}
	model, cmd := model.handleNewChatInputModeKey(tea.KeyMsg{Type: tea.KeyCtrlJ})
	if cmd == nil || !model.app.NewChatComposePending || !model.app.UserSearchLoading {
		t.Fatalf("Ctrl+Enter creation state cmd=%v pending=%v loading=%v", cmd != nil, model.app.NewChatComposePending, model.app.UserSearchLoading)
	}
}

func TestNewChatPickerOffersExactEmailWithoutDirectory(t *testing.T) {
	model := newWorkflowTestModel()
	model, _ = model.openNewChatPicker()
	model.app.UserSearchQuery = "new.person@example.com"
	model.userSearchInput.SetValue(model.app.UserSearchQuery)
	model.updateNewChatLocalResults()
	items := model.getUserSearchItems()
	if len(items) != 1 || items[0].DirUser == nil || items[0].Source != "Exact email" {
		t.Fatalf("exact-email fallback items = %#v", items)
	}
}

func TestNewChatCandidatesDeduplicateDirectoryAndExactEmail(t *testing.T) {
	model := newWorkflowTestModel()
	model.app.NewChatMode = true
	model.app.UserSearchQuery = "alice@example.com"
	model.app.NewChatLocalResults = []User{testDirectoryUser("alice-id", "Alice", "alice@example.com")}
	model.app.UserSearchDirectoryResults = []User{testDirectoryUser("alice-id", "Alice Example", "alice@example.com")}
	items := model.getUserSearchItems()
	if len(items) != 1 {
		t.Fatalf("duplicate participant candidates = %#v", items)
	}
}

func TestBuildChatCreatePayloadChoosesDirectOrGroup(t *testing.T) {
	alice := testDirectoryUser("alice-id", "Alice", "alice@example.com")
	bob := testDirectoryUser("bob-id", "Bob", "bob@example.com")
	direct, err := buildChatCreatePayload("me-id", []User{alice})
	if err != nil || direct["chatType"] != "oneOnOne" {
		t.Fatalf("direct payload = %#v err=%v", direct, err)
	}
	group, err := buildChatCreatePayload("me-id", []User{alice, bob})
	if err != nil || group["chatType"] != "group" {
		t.Fatalf("group payload = %#v err=%v", group, err)
	}
	if members, ok := group["members"].([]map[string]any); !ok || len(members) != 3 {
		t.Fatalf("group members = %#v", group["members"])
	}
}

func TestChatIDFromCreateLocation(t *testing.T) {
	want := "19:meeting_example@thread.v2"
	locations := []string{
		"https://graph.microsoft.com/v1.0/chats('19:meeting_example@thread.v2')/operations('op')",
		"https://graph.microsoft.com/v1.0/chats/19%3Ameeting_example%40thread.v2",
	}
	for _, location := range locations {
		if got := chatIDFromLocation(location); got != want {
			t.Fatalf("chatIDFromLocation(%q) = %q, want %q", location, got, want)
		}
	}
}
