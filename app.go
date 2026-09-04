package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// NotificationMode controls how the app notifies the user of new messages.
type NotificationMode int

const (
	NotificationNone    NotificationMode = iota // 0 — no notifications
	NotificationConsole                         // 1 — BEL + visual bell
	NotificationSystem                          // 2 — desktop notification
	NotificationBoth                            // 3 — BEL + visual bell + desktop
)

// String returns the human-readable name for a NotificationMode.
func (n NotificationMode) String() string {
	switch n {
	case NotificationConsole:
		return "Console"
	case NotificationSystem:
		return "System"
	case NotificationBoth:
		return "Both"
	default:
		return "None"
	}
}

// ChatReadFilter controls which server/local read states appear in the chat list.
type ChatReadFilter string

const (
	ChatReadAll    ChatReadFilter = "all"
	ChatReadUnread ChatReadFilter = "unread"
	ChatReadRead   ChatReadFilter = "read"
)

// ChatListFilter is the local, zero-network filter applied to the chat sidebar.
// An empty ChatTypes map means all chat types are included.
type ChatListFilter struct {
	Query          string
	ReadState      ChatReadFilter
	ChatTypes      map[string]bool
	FavouritesOnly bool
	TodayOnly      bool
	WithinHours    int
	SnoozedOnly    bool
}

func newChatListFilter() ChatListFilter {
	return ChatListFilter{
		ReadState: ChatReadAll,
		ChatTypes: make(map[string]bool),
	}
}

func cloneChatListFilter(filter ChatListFilter) ChatListFilter {
	clone := filter
	clone.ChatTypes = make(map[string]bool, len(filter.ChatTypes))
	for chatType, enabled := range filter.ChatTypes {
		clone.ChatTypes[chatType] = enabled
	}
	return clone
}

// MarshalJSON serialises NotificationMode as a string so config.json is readable.
func (n NotificationMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

// UnmarshalJSON parses a NotificationMode from its string representation.
func (n *NotificationMode) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		// Also accept integer representation for backward compatibility.
		var i int
		if err2 := json.Unmarshal(data, &i); err2 != nil {
			return err
		}
		*n = NotificationMode(i)
		return nil
	}
	switch s {
	case "Console":
		*n = NotificationConsole
	case "System":
		*n = NotificationSystem
	case "Both":
		*n = NotificationBoth
	default:
		*n = NotificationNone
	}
	return nil
}

// ---------------------------------------------------------------------------
// App — central application state
// ---------------------------------------------------------------------------

// FeatureFlags holds resolved optional feature flags for the running session.
// Populated once at startup from config to avoid repeated file reads.
type FeatureFlags struct {
	FilePreview           bool // requires Files.Read
	FilePreviewInTerminal bool // show image in terminal if FilePreview is enabled
	FileUpload            bool // requires Files.ReadWrite
	Presence              bool // requires Presence.Read.All
	UserProfile           bool // requires User.ReadBasic.All (or User.Read.All if Extended)
	ProfileExtended       bool // requires User.Read.All (admin consent)
	TeamsChannels         bool // requires Team.ReadBasic.All + Channel.ReadBasic.All
	ChannelMentions       bool // requires TeamMember.Read.All
	SqliteEnabled         bool
}

// App holds all runtime state for the Teams TUI application.
type App struct {
	Chats                      []Chat
	Status                     string
	SearchStatus               string
	SelectedIndex              int    // derived visible-row cursor for rendering
	SelectedChatID             string // canonical selected chat identity
	CurrentUserName            *string
	CurrentUserID              string // used for markChatRead
	Messages                   []Message
	MessagesConversationID     string // chat/channel that owns Messages
	LoadingMessages            bool
	SearchLoadingMessages      bool
	InputMode                  bool
	InputBuffer                string
	ScrollOffset               int
	MaxScroll                  int
	ChatScrollOffset           int
	ChannelScrollOffset        int
	SnapToBottom               bool
	MessageSelectedIndex       int
	MessageSelectionMode       bool
	MessagePopupMode           bool
	MessagePopupScrollOffset   int
	ReactionMode               bool
	DeleteConfirmMode          bool
	NotificationMode           NotificationMode
	NotificationShowPreview    bool
	NotificationPreviewLen     int
	MarkReadOnOpen             bool
	DefaultSnoozeMinutes       int
	WorkdayStart               string
	WorkdayEnd                 string
	ExportDirectory            string
	ThreadAnalysisAgent        string
	ThreadAnalysisCommand      string
	ThreadCaptureFormat        ThreadCaptureFormat
	ThreadCaptureFile          string
	ThreadCaptureOrgFile       string
	ShowChatDates              bool
	VisualBellUntil            *time.Time
	StatusUntil                *time.Time
	SearchStatusUntil          *time.Time
	NextLink                   string
	PendingScrollID            string
	EditingMessageID           *string
	ReplyToMessage             *Message // source message for a quoted reply
	PendingForwardText         string   // editable Markdown prepared while choosing a forward target
	UrlSelectionMode           bool
	UrlSelectionOpenMode       bool // true if opening, false if yanking/copying
	UrlSelectedIndex           int
	UrlsInMessage              []string
	MessageLineOffsets         []int
	SearchMode                 bool
	SearchActive               bool
	SearchQuery                string
	SearchPopupMode            bool
	SearchPopupSelectedIndex   int
	SearchPopupScrollOffset    int
	SearchPopupResults         []SearchPopupItem
	HistoryMessages            map[string][]Message
	HistoryNextLink            map[string]string
	HistoryInitialized         map[string]bool
	ChatMessagesLoadedOnce     map[string]bool
	ChatCacheDirty             map[string]bool
	SearchStates               map[string]*ChatSearchState
	CachedMessages             map[string][]Message // per-chat message cache for instant restore on revisit
	CachedNextLink             map[string]string    // per-chat NextLink cache
	MainChatScrollOffset       int
	MainChatSnapToBottom       bool
	UserSearchPopupMode        bool
	UserSearchMode             bool
	UserSearchQuery            string
	UserSearchStatus           string
	UserSearchStatusUntil      *time.Time
	UserSearchLocalResults     []Chat
	UserSearchMemberResults    []Chat
	UserSearchMessageResults   []MessageSearchResult
	UserSearchChannelResults   []channelEntry
	UserSearchDirectoryResults []User
	UserSearchSelectedIndex    int
	UserSearchLoading          bool
	NewChatMode                bool
	NewChatLocalResults        []User
	NewChatSelectedUsers       []User
	NewChatDirectoryQuery      string
	NewChatComposePending      bool
	ChatFilterPopupMode        bool
	ChatFilterInputMode        bool
	ChatFilterSelectedIndex    int
	ChatBookmarkPopupMode      bool
	ChatBookmarkSelectedIndex  int
	ChatBookmarks              []ChatBookmarkConfig
	ActiveChatBookmark         string
	ThreadActionPopupMode      bool
	ThreadActionSelectedIndex  int
	ArtifactPopupMode          bool
	ArtifactSelectedIndex      int
	Artifacts                  []ConversationArtifact
	ActiveChatFilter           ChatListFilter
	DraftChatFilter            ChatListFilter
	UnreadOverlay              bool
	SnoozePopupMode            bool
	SnoozeSelectedIndex        int
	AppStartTime               time.Time
	ChatIconTheme              string
	CustomChatIcons            map[string]string
	Features                   FeatureFlags

	// ── Presence popup (Feature: presence_enabled) ───────────────────────
	PresencePopupMode    bool
	PresenceChatMode     bool
	PresenceData         *UserPresence
	PresenceChatData     []PresenceEntry
	PresenceUserName     string // display name of the user whose presence is shown
	PresenceLoading      bool
	PresenceScrollOffset int

	// ── User Profile popup (Feature: user_profile_enabled) ───────────────
	UserProfilePopupMode bool
	UserProfileData      *UserProfile
	UserProfileLoading   bool

	// ── Attachment cursor in message view popup (Feature: file_preview_enabled) ──
	AttachmentSelectedIndex int
	AttachmentCursorMode    bool // true when navigating attachments inside the v popup

	// ── Teams channels data (Feature: teams_channels_enabled) ───────────
	TeamsData             []TeamWithChannels // cached joined teams + channels; populated at startup
	TeamsDataLoading      bool
	SelectedChannelTeamID string // teamID of the currently viewed channel ("" = chat mode)
	SelectedChannelID     string // channelID of the currently viewed channel ("" = chat mode)
	ChannelReplyToID      string // root message ID when replying to a channel thread ("" = new root post)
	ChannelMsgRefreshMin  int
	ExternalEditor        string // command/path for the external editor
	ConfigPath            string // effective config.json path for diagnostics
	BrowserCommand        string // command to open URLs
	TeamsAppCommand       string // command to open msteams:// desktop deep links
	ImageViewer           string // command to open images (empty = use default file opener)
	YoutrackCommand       string // command to open YouTrack URLs
	GitlabCommand         string // command to open GitLab URLs

	// ── Mention Popup Autocomplete ───────────────────────────────────────
	MentionPopupMode          bool
	MentionSearch             string
	MentionSelectedIndex      int
	MentionScrollOffset       int
	MentionSuggestions        []ChatMember
	MentionStartIndex         int
	MentionCanceledStartIndex int
	TeamMembersCache          map[string][]ChatMember

	// ── Help popup ───────────────────────────────────────────────────────
	HelpPopupMode    bool
	HelpScrollOffset int

	// ── File Picker popup ────────────────────────────────────────────────
	FilePickerPopupMode bool

	// ── Composed images (pasted from clipboard) ──────────────────────────
	ComposedImages     []PastedImage
	ComposedFiles      []PendingFile
	SkipTextareaUpdate bool
}

// PendingFile represents a file selected from the file system.
type PendingFile struct {
	Name        string
	Path        string
	Data        []byte
	ContentType string
}

// ChatSearchState holds the search-specific query and viewport navigation state for a chat.
type ChatSearchState struct {
	Query           string
	Results         []SearchPopupItem
	SelectedIndex   int
	ScrollOffset    int
	ExpandedIndices map[int]bool
	Status          string
}

// SearchPopupItem represents a message displayed inside the search popup (with context flag).
type SearchPopupItem struct {
	Message      Message
	IsMatch      bool
	HistoryIndex int
}

// NewApp creates an App with sensible initial defaults.
func NewApp() *App {
	defaultFilter := newChatListFilter()
	return &App{
		Status:                    "Loading...",
		SnapToBottom:              true,
		NotificationMode:          NotificationNone,
		NotificationShowPreview:   false,
		NotificationPreviewLen:    50,
		HistoryMessages:           make(map[string][]Message),
		HistoryNextLink:           make(map[string]string),
		HistoryInitialized:        make(map[string]bool),
		ChatMessagesLoadedOnce:    make(map[string]bool),
		ChatCacheDirty:            make(map[string]bool),
		SearchStates:              make(map[string]*ChatSearchState),
		CachedMessages:            make(map[string][]Message),
		CachedNextLink:            make(map[string]string),
		ActiveChatFilter:          defaultFilter,
		DraftChatFilter:           cloneChatListFilter(defaultFilter),
		TeamMembersCache:          make(map[string][]ChatMember),
		ChatIconTheme:             "unicode",
		CustomChatIcons:           make(map[string]string),
		AppStartTime:              time.Now(),
		MentionCanceledStartIndex: -1,
		DefaultSnoozeMinutes:      180,
		WorkdayStart:              "07:00",
		WorkdayEnd:                "18:00",
	}
}

// SetChats replaces the chat list and updates the status line.
func (a *App) SetChats(chats []Chat) {
	a.Chats = chats
	a.SyncSelectedChat()
	a.SetStatus(fmt.Sprintf("Loaded %d chats", len(chats)), 5*time.Second)
}

// SetCurrentUser records the detected current user's display name.
func (a *App) SetCurrentUser(name string) {
	a.CurrentUserName = &name
}

// ActivateMessagesConversation makes conversationID the owner of the right
// pane. Switching owners clears transient state so messages from two chats can
// never be displayed or merged together.
func (a *App) ActivateMessagesConversation(conversationID string) {
	if a.MessagesConversationID == conversationID && messagesMatchConversation(a.Messages, conversationID) {
		return
	}
	a.MessagesConversationID = conversationID
	a.Messages = nil
	a.NextLink = ""
	a.PendingScrollID = ""
	a.ScrollOffset = 0
	a.MaxScroll = 0
	a.SnapToBottom = true
	a.MessageSelectedIndex = 0
	a.MessageSelectionMode = false
	a.MessagePopupMode = false
	a.AttachmentCursorMode = false
	a.AttachmentSelectedIndex = 0
}

// ClearMessagesConversation removes the right-pane owner and transcript.
func (a *App) ClearMessagesConversation() {
	a.MessagesConversationID = ""
	a.Messages = nil
	a.NextLink = ""
	a.PendingScrollID = ""
	a.ScrollOffset = 0
	a.MaxScroll = 0
	a.SnapToBottom = true
	a.MessageSelectedIndex = 0
	a.MessageSelectionMode = false
	a.MessagePopupMode = false
	a.AttachmentCursorMode = false
	a.AttachmentSelectedIndex = 0
	a.SetLoadingMessages(false)
}

// MessagesBelongTo reports whether the right-pane transcript belongs to the
// given chat or channel.
func (a *App) MessagesBelongTo(conversationID string) bool {
	return conversationID != "" && a.MessagesConversationID == conversationID &&
		messagesMatchConversation(a.Messages, conversationID)
}

func messagesMatchConversation(messages []Message, conversationID string) bool {
	if conversationID == "" {
		return len(messages) == 0
	}
	for _, message := range messages {
		if message.ChatID != "" && message.ChatID != conversationID {
			return false
		}
		if message.ChannelIdentity != nil && message.ChannelIdentity.ChannelID != "" &&
			message.ChannelIdentity.ChannelID != conversationID {
			return false
		}
	}
	return true
}

// SetMessages updates one conversation's message list. A response for a new
// owner replaces the pane; only responses for the same owner are merged.
func (a *App) SetMessages(conversationID string, messages []Message, nextLink string) {
	a.ActivateMessagesConversation(conversationID)
	if len(a.Messages) == 0 {
		a.Messages = append([]Message(nil), messages...)
		sortMessagesNewestFirst(a.Messages)
		a.NextLink = nextLink
		a.LoadingMessages = false
		return
	}

	// Create a map of existing messages by ID.
	m := make(map[string]Message)
	for _, msg := range a.Messages {
		m[msg.ID] = msg
	}
	// Overwrite/add with fresh messages.
	for _, msg := range messages {
		m[msg.ID] = msg
	}

	result := make([]Message, 0, len(m))
	for _, msg := range m {
		result = append(result, msg)
	}

	// Sort newest first using absolute RFC 3339 timestamps.
	sortMessagesNewestFirst(result)

	// Maintain selection by ID if in message selection mode.
	if a.MessageSelectionMode && len(a.Messages) > 0 && a.MessageSelectedIndex < len(a.Messages) {
		selectedID := a.Messages[a.MessageSelectedIndex].ID
		a.Messages = result // set here so we can find index in new list
		for i, m := range result {
			if m.ID == selectedID {
				a.MessageSelectedIndex = i
				goto done
			}
		}
	}

	a.Messages = result

done:
	// Clamp index if message was deleted or we're out of bounds.
	if a.MessageSelectedIndex >= len(a.Messages) {
		if len(a.Messages) > 0 {
			a.MessageSelectedIndex = len(a.Messages) - 1
		} else {
			a.MessageSelectedIndex = 0
		}
	}

	// Only update NextLink if it's currently empty (e.g. first successful load).
	// If we already have a NextLink, it points to the older history we've reached.
	if a.NextLink == "" {
		a.NextLink = nextLink
	}
	a.LoadingMessages = false
}

// AppendOlderMessages adds an older page only while the same conversation
// still owns the right pane. It returns false for a stale page.
func (a *App) AppendOlderMessages(conversationID string, messages []Message, nextLink string) bool {
	if !a.MessagesBelongTo(conversationID) {
		return false
	}
	a.Messages = append(a.Messages, messages...)
	sortMessagesNewestFirst(a.Messages)
	a.NextLink = nextLink
	a.LoadingMessages = false
	return true
}

// SetLoadingMessages toggles the loading indicator.
func (a *App) SetLoadingMessages(loading bool) {
	a.LoadingMessages = loading
}

// SetSearchLoadingMessages toggles the search loading indicator.
func (a *App) SetSearchLoadingMessages(loading bool) {
	a.SearchLoadingMessages = loading
}

// SetSearchStatus sets the search status text, optionally clearing it after duration.
func (a *App) SetSearchStatus(msg string, duration time.Duration) {
	a.SearchStatus = msg
	if duration > 0 {
		t := time.Now().Add(duration)
		a.SearchStatusUntil = &t
	} else {
		a.SearchStatusUntil = nil
	}
}

// SyncSelectedChat reconciles the derived row cursor with the canonical chat
// identity. An existing identity always wins over a stale numeric index.
func (a *App) SyncSelectedChat() bool {
	if a.SelectedChatID != "" {
		for index := range a.Chats {
			if a.Chats[index].ID == a.SelectedChatID {
				a.SelectedIndex = index
				return true
			}
		}
		a.ClearSelectedChat()
		return false
	}
	if a.SelectedIndex >= 0 && a.SelectedIndex < len(a.Chats) {
		a.SelectedChatID = a.Chats[a.SelectedIndex].ID
		return true
	}
	a.SelectedIndex = -1
	return false
}

// SetSelectedChatIndex selects one visible row and records its stable identity.
func (a *App) SetSelectedChatIndex(index int) bool {
	if index < 0 || index >= len(a.Chats) {
		a.ClearSelectedChat()
		return false
	}
	a.SelectedIndex = index
	a.SelectedChatID = a.Chats[index].ID
	return true
}

// SetSelectedChatID selects a visible chat by identity.
func (a *App) SetSelectedChatID(chatID string) bool {
	if chatID == "" {
		a.ClearSelectedChat()
		return false
	}
	for index := range a.Chats {
		if a.Chats[index].ID == chatID {
			a.SelectedChatID = chatID
			a.SelectedIndex = index
			return true
		}
	}
	return false
}

// ClearSelectedChat returns the chat sidebar to its unselected dashboard state.
func (a *App) ClearSelectedChat() {
	a.SelectedChatID = ""
	a.SelectedIndex = -1
}

// GetSelectedChat returns the currently highlighted chat, or nil.
func (a *App) GetSelectedChat() *Chat {
	if !a.SyncSelectedChat() {
		return nil
	}
	return &a.Chats[a.SelectedIndex]
}

// NextChat moves the selection one step down, wrapping around.
func (a *App) NextChat() {
	if len(a.Chats) == 0 {
		return
	}
	index := a.SelectedIndex
	if !a.SyncSelectedChat() {
		index = -1
	} else {
		index = a.SelectedIndex
	}
	a.SetSelectedChatIndex((index + 1) % len(a.Chats))
}

// PreviousChat moves the selection one step up, wrapping around.
func (a *App) PreviousChat() {
	if len(a.Chats) == 0 {
		return
	}
	if !a.SyncSelectedChat() {
		a.SetSelectedChatIndex(len(a.Chats) - 1)
		return
	}
	a.SetSelectedChatIndex((a.SelectedIndex - 1 + len(a.Chats)) % len(a.Chats))
}

// ToggleNotificationMode cycles None → Console → System → Both → None.
func (a *App) ToggleNotificationMode() {
	a.NotificationMode = (a.NotificationMode + 1) % 4
}

// TriggerVisualBell sets VisualBellUntil to 200 ms from now.
func (a *App) TriggerVisualBell() {
	t := time.Now().Add(200 * time.Millisecond)
	a.VisualBellUntil = &t
}

// VisualBellActive reports whether the visual bell should be showing.
func (a *App) VisualBellActive() bool {
	return a.VisualBellUntil != nil && time.Now().Before(*a.VisualBellUntil)
}

// SetStatus sets the status line, optionally clearing it after duration.
func (a *App) SetStatus(msg string, duration time.Duration) {
	a.Status = msg
	if duration > 0 {
		t := time.Now().Add(duration)
		a.StatusUntil = &t
	} else {
		a.StatusUntil = nil
	}
}
