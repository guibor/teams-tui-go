package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// KeyList accepts either a single JSON string or an array of strings. An empty
// list explicitly unbinds an action, while an omitted action keeps its default.
type KeyList []string

func (keys *KeyList) UnmarshalJSON(data []byte) error {
	var one string
	if err := json.Unmarshal(data, &one); err == nil {
		*keys = KeyList{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return fmt.Errorf("key binding must be a string or array of strings: %w", err)
	}
	*keys = KeyList(many)
	return nil
}

// KeyBindingConfig maps stable action names to one or more terminal key names.
type KeyBindingConfig map[string]KeyList

type keyContext string

const (
	keyContextNormalChat     keyContext = "normal-chat"
	keyContextNormalChannel  keyContext = "normal-channel"
	keyContextCompose        keyContext = "compose"
	keyContextMention        keyContext = "mention"
	keyContextMessageSelect  keyContext = "message-select"
	keyContextMessageView    keyContext = "message-view"
	keyContextSearchInput    keyContext = "search-input"
	keyContextSearchResults  keyContext = "search-results"
	keyContextGlobalInput    keyContext = "global-input"
	keyContextGlobalResults  keyContext = "global-results"
	keyContextNewChatInput   keyContext = "new-chat-input"
	keyContextNewChatResults keyContext = "new-chat-results"
	keyContextFilterInput    keyContext = "filter-input"
	keyContextFilter         keyContext = "filter"
	keyContextBookmarks      keyContext = "bookmarks"
	keyContextThreadActions  keyContext = "thread-actions"
	keyContextArtifacts      keyContext = "artifacts"
	keyContextHelp           keyContext = "help"
	keyContextURLList        keyContext = "url-list"
	keyContextReaction       keyContext = "reaction"
	keyContextDeleteConfirm  keyContext = "delete-confirm"
	keyContextPresence       keyContext = "presence"
	keyContextProfile        keyContext = "profile"
)

const (
	keyAppQuitNow             = "app.quit_now"
	keyAppQuit                = "app.quit"
	keyDashboardLeave         = "dashboard.leave"
	keyChatFirst              = "chat.first"
	keyChatFirstPrefix        = "chat.first_prefix"
	keyChatLast               = "chat.last"
	keyChannelFirst           = "channel.first"
	keyChannelLast            = "channel.last"
	keyChatNext               = "chat.next"
	keyChatPrevious           = "chat.previous"
	keySidebarSwitch          = "sidebar.switch_section"
	keyNotificationsCycle     = "notifications.cycle"
	keyChatDatesToggle        = "chat_dates.toggle"
	keyHelpOpen               = "help.open"
	keyComposeStart           = "compose.start"
	keyNewChatOpen            = "new_chat.open"
	keySearchGlobal           = "search.global"
	keyFilterOpen             = "filter.open"
	keyUnreadOverlay          = "filter.toggle_unread_overlay"
	keyBookmarksOpen          = "bookmarks.open"
	keyThreadActionsOpen      = "thread_actions.open"
	keyArtifactsOpen          = "artifacts.open"
	keySearchHistory          = "search.history"
	keyMessagePageUp          = "message.page_up"
	keyMessagePageDown        = "message.page_down"
	keyMessageTop             = "message.top"
	keyMessageBottom          = "message.bottom"
	keyMessageSelectMode      = "message.select_mode"
	keyMessageSelectClose     = "message.exit_select_mode"
	keyChatFavorite           = "chat.toggle_favorite"
	keyChatOpenBrowser        = "chat.open_browser"
	keyChatOpenApp            = "chat.open_desktop_app"
	keyMessageReply           = "message.reply"
	keyMessageForward         = "message.forward"
	keyChatMarkRead           = "chat.mark_read"
	keyChatMarkUnread         = "chat.mark_unread"
	keyChatExport             = "chat.export_markdown"
	keyChatAnalyze            = "chat.analyze"
	keyPresenceOpen           = "presence.open"
	keyChannelToggleHidden    = "channel.toggle_hidden"
	keyMessageNext            = "message.next"
	keyMessagePrevious        = "message.previous"
	keyMessageNewest          = "message.newest"
	keyMessageOldest          = "message.oldest"
	keyMessageView            = "message.view"
	keyMessageCopy            = "message.copy"
	keyMessageCopyURLs        = "message.copy_urls"
	keyMessageOpenURLs        = "message.open_urls"
	keyMessageReact           = "message.react"
	keyMessageDelete          = "message.delete"
	keyMessageEdit            = "message.edit"
	keyMessageProfile         = "message.profile"
	keyMessageExternalEditor  = "message.external_editor"
	keyMessageViewClose       = "message_view.close"
	keyMessageViewAttachments = "message_view.toggle_attachments"
	keyMessageViewOpen        = "message_view.open_attachment"
	keyMessageViewScrollDown  = "message_view.scroll_down"
	keyMessageViewScrollUp    = "message_view.scroll_up"
	keyComposeCancel          = "compose.cancel"
	keyComposeExternalEditor  = "compose.external_editor"
	keyComposeAttach          = "compose.attach_file"
	keyComposePasteImage      = "compose.paste_image"
	keyComposeSendPrefix      = "compose.send_prefix"
	keyComposeSend            = "compose.send"
	keyComposeNewline         = "compose.newline"
	keyMentionClose           = "mention.close"
	keyMentionPrevious        = "mention.previous"
	keyMentionNext            = "mention.next"
	keyMentionSelect          = "mention.select"
	keyInputCancel            = "input.cancel"
	keyInputSubmit            = "input.submit"
	keyInputFocusResults      = "input.focus_results"
	keyListClose              = "list.close"
	keyListNext               = "list.next"
	keyListPrevious           = "list.previous"
	keyListSelect             = "list.select"
	keyListEditQuery          = "list.edit_query"
	keySearchJump             = "search_result.jump"
	keySearchCopyMessage      = "search_result.copy_message"
	keySearchCopyURLs         = "search_result.copy_urls"
	keySearchOpenURLs         = "search_result.open_urls"
	keyNewChatCreate          = "new_chat.create"
	keyNewChatCancelInput     = "new_chat.cancel_input"
	keyNewChatToggleFirst     = "new_chat.toggle_first_participant"
	keyNewChatToggle          = "new_chat.toggle_participant"
	keyFilterToggle           = "filter.toggle_selected"
	keyFilterUnread           = "filter.unread"
	keyFilterRead             = "filter.read"
	keyFilterAll              = "filter.all_read_states"
	keyFilterToday            = "filter.today"
	keyFilterDirect           = "filter.direct"
	keyFilterGroup            = "filter.group"
	keyFilterMeeting          = "filter.meeting"
	keyFilterFavorites        = "filter.favorites"
	keyFilterClear            = "filter.clear"
	keyBookmarkClose          = "bookmarks.close"
	keyThreadOpenBrowser      = "thread_action.open_browser"
	keyThreadOpenApp          = "thread_action.open_desktop_app"
	keyThreadCompose          = "thread_action.compose"
	keyThreadReply            = "thread_action.reply"
	keyThreadForward          = "thread_action.forward"
	keyThreadRead             = "thread_action.mark_read"
	keyThreadUnread           = "thread_action.mark_unread"
	keyThreadFavorite         = "thread_action.toggle_favorite"
	keyThreadCapture          = "thread_action.capture"
	keyThreadExport           = "thread_action.export_markdown"
	keyThreadAnalyze          = "thread_action.analyze"
	keyThreadCopyLink         = "thread_action.copy_link"
	keyThreadArtifacts        = "thread_action.artifacts"
	keyArtifactOpen           = "artifact.open"
	keyArtifactCopy           = "artifact.copy_link"
	keyHelpClose              = "help.close"
	keyHelpDown               = "help.scroll_down"
	keyHelpUp                 = "help.scroll_up"
	keyURLCopy                = "url.copy"
	keyURLOpen                = "url.open"
	keyReactionClose          = "reaction.close"
	keyReactionLike           = "reaction.like"
	keyReactionHeart          = "reaction.heart"
	keyReactionLaugh          = "reaction.laugh"
	keyReactionSurprised      = "reaction.surprised"
	keyReactionSad            = "reaction.sad"
	keyReactionAngry          = "reaction.angry"
	keyDeleteConfirm          = "delete.confirm"
	keyDeleteCancel           = "delete.cancel"
	keyPresenceClose          = "presence.close"
	keyPresencePrevious       = "presence.previous"
	keyPresenceNext           = "presence.next"
	keyProfileClose           = "profile.close"
)

type keyBindingDefinition struct {
	Action    string
	Canonical string
	Defaults  KeyList
	Contexts  []keyContext
}

func keyDef(action, canonical string, defaults KeyList, contexts ...keyContext) keyBindingDefinition {
	return keyBindingDefinition{Action: action, Canonical: canonical, Defaults: defaults, Contexts: contexts}
}

var keyBindingDefinitions = []keyBindingDefinition{
	keyDef(keyAppQuitNow, "ctrl+c", KeyList{"ctrl+c"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyAppQuit, "q", KeyList{"q"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyDashboardLeave, "esc", KeyList{"esc"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyChatFirst, "alt+<", KeyList{"alt+<", "h"}, keyContextNormalChat),
	keyDef(keyChatFirstPrefix, "g", KeyList{"g"}, keyContextNormalChat),
	keyDef(keyChatLast, "alt+>", KeyList{"alt+>", "G", "l"}, keyContextNormalChat),
	keyDef(keyChannelFirst, "alt+<", KeyList{"alt+<"}, keyContextNormalChannel),
	keyDef(keyChannelLast, "alt+>", KeyList{"alt+>"}, keyContextNormalChannel),
	keyDef(keyChatNext, "j", KeyList{"j", "down", "alt+n"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyChatPrevious, "k", KeyList{"k", "up", "alt+p"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keySidebarSwitch, "tab", KeyList{"tab"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyNotificationsCycle, "n", KeyList{"n"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyChatDatesToggle, "D", KeyList{"D"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyHelpOpen, "?", KeyList{"?"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyComposeStart, "c", KeyList{"c", "C"}, keyContextNormalChat, keyContextNormalChannel, keyContextMessageSelect, keyContextMessageView),
	keyDef(keyNewChatOpen, "N", KeyList{"N"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keySearchGlobal, "s", KeyList{"s"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyFilterOpen, "v", KeyList{"v", "V"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyUnreadOverlay, "U", KeyList{"U"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyBookmarksOpen, "b", KeyList{"b"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyThreadActionsOpen, "a", KeyList{"a"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyArtifactsOpen, "T", KeyList{"T"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keySearchHistory, "/", KeyList{"/"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyMessagePageUp, "K", KeyList{"K", "pgup"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyMessagePageDown, "J", KeyList{"J", "pgdown"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyMessageTop, "<", KeyList{"<", "H"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyMessageBottom, ">", KeyList{">", "L"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyMessageSelectMode, "m", KeyList{"m"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyMessageSelectClose, "esc", KeyList{"esc", "m"}, keyContextMessageSelect),
	keyDef(keyChatFavorite, "*", KeyList{"*"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyChatOpenBrowser, "o", KeyList{"o"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyChatOpenApp, "O", KeyList{"O"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyMessageReply, "R", KeyList{"R"}, keyContextNormalChat, keyContextNormalChannel, keyContextMessageSelect, keyContextMessageView),
	keyDef(keyMessageForward, "f", KeyList{"f", "F"}, keyContextNormalChat, keyContextNormalChannel, keyContextMessageSelect, keyContextMessageView),
	keyDef(keyChatMarkRead, "r", KeyList{"r", "i"}, keyContextNormalChat, keyContextNormalChannel, keyContextMessageSelect, keyContextMessageView),
	keyDef(keyChatMarkUnread, "u", KeyList{"u"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyChatExport, "E", KeyList{"E"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyChatAnalyze, "A", KeyList{"A"}, keyContextNormalChat, keyContextNormalChannel),
	keyDef(keyPresenceOpen, "p", KeyList{"p"}, keyContextNormalChat, keyContextNormalChannel, keyContextMessageSelect),
	keyDef(keyChannelToggleHidden, "h", KeyList{"h"}, keyContextNormalChannel),

	keyDef(keyComposeCancel, "esc", KeyList{"esc"}, keyContextCompose),
	keyDef(keyComposeExternalEditor, "ctrl+g", KeyList{"ctrl+g"}, keyContextCompose),
	keyDef(keyComposeAttach, "ctrl+f", KeyList{"ctrl+f"}, keyContextCompose),
	keyDef(keyComposePasteImage, "ctrl+v", KeyList{"ctrl+v", "ctrl+shift+v"}, keyContextCompose),
	keyDef(keyComposeSendPrefix, "ctrl+c", KeyList{"ctrl+c"}, keyContextCompose),
	keyDef(keyComposeSend, "ctrl+j", KeyList{"ctrl+j", "ctrl+enter"}, keyContextCompose),
	keyDef(keyComposeNewline, "enter", KeyList{"enter", "alt+enter", "shift+enter"}, keyContextCompose),
	keyDef(keyMentionClose, "esc", KeyList{"esc"}, keyContextMention),
	keyDef(keyMentionPrevious, "up", KeyList{"up", "shift+tab"}, keyContextMention),
	keyDef(keyMentionNext, "down", KeyList{"down", "tab"}, keyContextMention),
	keyDef(keyMentionSelect, "enter", KeyList{"enter"}, keyContextMention),

	keyDef(keyMessageNext, "j", KeyList{"j", "down"}, keyContextMessageSelect, keyContextMessageView),
	keyDef(keyMessagePrevious, "k", KeyList{"k", "up"}, keyContextMessageSelect, keyContextMessageView),
	keyDef(keyMessageNewest, ">", KeyList{">", "L"}, keyContextMessageSelect, keyContextMessageView),
	keyDef(keyMessageOldest, "<", KeyList{"<", "H"}, keyContextMessageSelect, keyContextMessageView),
	keyDef(keyMessageView, "v", KeyList{"v"}, keyContextMessageSelect),
	keyDef(keyMessageCopy, "y", KeyList{"y"}, keyContextMessageSelect),
	keyDef(keyMessageCopyURLs, "u", KeyList{"u"}, keyContextMessageSelect),
	keyDef(keyMessageOpenURLs, "o", KeyList{"o"}, keyContextMessageSelect),
	keyDef(keyMessageReact, "+", KeyList{"+", "a"}, keyContextMessageSelect),
	keyDef(keyMessageDelete, "d", KeyList{"d"}, keyContextMessageSelect),
	keyDef(keyMessageEdit, "e", KeyList{"e"}, keyContextMessageSelect),
	keyDef(keyMessageProfile, "I", KeyList{"I"}, keyContextMessageSelect),
	keyDef(keyMessageExternalEditor, "ctrl+g", KeyList{"ctrl+g"}, keyContextMessageSelect, keyContextMessageView),
	keyDef(keyMessageViewClose, "esc", KeyList{"esc", "q", "v"}, keyContextMessageView),
	keyDef(keyMessageViewAttachments, "tab", KeyList{"tab"}, keyContextMessageView),
	keyDef(keyMessageViewOpen, "enter", KeyList{"enter"}, keyContextMessageView),
	keyDef(keyMessageViewScrollDown, "J", KeyList{"J", "shift+down", "pgdown"}, keyContextMessageView),
	keyDef(keyMessageViewScrollUp, "K", KeyList{"K", "shift+up", "pgup"}, keyContextMessageView),

	keyDef(keyInputCancel, "esc", KeyList{"esc"}, keyContextSearchInput, keyContextGlobalInput, keyContextFilterInput),
	keyDef(keyInputSubmit, "enter", KeyList{"enter"}, keyContextSearchInput, keyContextGlobalInput, keyContextFilterInput),
	keyDef(keyInputFocusResults, "down", KeyList{"down", "up", "tab"}, keyContextGlobalInput, keyContextNewChatInput),
	keyDef(keyListClose, "esc", KeyList{"esc", "q"}, keyContextSearchResults, keyContextGlobalResults, keyContextNewChatResults, keyContextFilter, keyContextThreadActions, keyContextArtifacts, keyContextURLList),
	keyDef(keyListNext, "j", KeyList{"j", "down", "tab"}, keyContextSearchResults, keyContextGlobalResults, keyContextNewChatResults, keyContextFilter, keyContextBookmarks, keyContextThreadActions, keyContextArtifacts, keyContextURLList),
	keyDef(keyListPrevious, "k", KeyList{"k", "up", "shift+tab"}, keyContextSearchResults, keyContextGlobalResults, keyContextNewChatResults, keyContextFilter, keyContextBookmarks, keyContextThreadActions, keyContextArtifacts, keyContextURLList),
	keyDef(keyListSelect, "enter", KeyList{"enter"}, keyContextSearchResults, keyContextGlobalResults, keyContextFilter, keyContextBookmarks, keyContextThreadActions, keyContextURLList),
	keyDef(keyListEditQuery, "/", KeyList{"/"}, keyContextSearchResults, keyContextGlobalResults, keyContextNewChatResults, keyContextFilter),
	keyDef(keySearchJump, "g", KeyList{"g"}, keyContextSearchResults),
	keyDef(keySearchCopyMessage, "y", KeyList{"y"}, keyContextSearchResults),
	keyDef(keySearchCopyURLs, "u", KeyList{"u"}, keyContextSearchResults),
	keyDef(keySearchOpenURLs, "o", KeyList{"o"}, keyContextSearchResults),
	keyDef(keyNewChatCreate, "ctrl+j", KeyList{"ctrl+j", "ctrl+enter"}, keyContextNewChatInput, keyContextNewChatResults),
	keyDef(keyNewChatCancelInput, "esc", KeyList{"esc"}, keyContextNewChatInput),
	keyDef(keyNewChatToggleFirst, "enter", KeyList{"enter"}, keyContextNewChatInput),
	keyDef(keyNewChatToggle, "enter", KeyList{"enter", "space"}, keyContextNewChatResults),

	keyDef(keyFilterToggle, "space", KeyList{"space"}, keyContextFilter),
	keyDef(keyFilterUnread, "u", KeyList{"u"}, keyContextFilter),
	keyDef(keyFilterRead, "r", KeyList{"r"}, keyContextFilter),
	keyDef(keyFilterAll, "a", KeyList{"a"}, keyContextFilter),
	keyDef(keyFilterToday, "t", KeyList{"t"}, keyContextFilter),
	keyDef(keyFilterDirect, "1", KeyList{"1"}, keyContextFilter),
	keyDef(keyFilterGroup, "g", KeyList{"g"}, keyContextFilter),
	keyDef(keyFilterMeeting, "m", KeyList{"m"}, keyContextFilter),
	keyDef(keyFilterFavorites, "f", KeyList{"f"}, keyContextFilter),
	keyDef(keyFilterClear, "x", KeyList{"x"}, keyContextFilter),
	keyDef(keyBookmarkClose, "esc", KeyList{"esc", "q", "b"}, keyContextBookmarks),

	keyDef(keyThreadOpenBrowser, "o", KeyList{"o"}, keyContextThreadActions),
	keyDef(keyThreadOpenApp, "O", KeyList{"O"}, keyContextThreadActions),
	keyDef(keyThreadCompose, "c", KeyList{"c", "C"}, keyContextThreadActions),
	keyDef(keyThreadReply, "R", KeyList{"R"}, keyContextThreadActions),
	keyDef(keyThreadForward, "f", KeyList{"f", "F"}, keyContextThreadActions),
	keyDef(keyThreadRead, "r", KeyList{"r", "i"}, keyContextThreadActions),
	keyDef(keyThreadUnread, "u", KeyList{"u"}, keyContextThreadActions),
	keyDef(keyThreadFavorite, "*", KeyList{"*"}, keyContextThreadActions),
	keyDef(keyThreadCapture, "a", KeyList{"a"}, keyContextThreadActions),
	keyDef(keyThreadExport, "e", KeyList{"e"}, keyContextThreadActions),
	keyDef(keyThreadAnalyze, "A", KeyList{"A"}, keyContextThreadActions),
	keyDef(keyThreadCopyLink, "y", KeyList{"y"}, keyContextThreadActions),
	keyDef(keyThreadArtifacts, "t", KeyList{"t"}, keyContextThreadActions),
	keyDef(keyArtifactOpen, "enter", KeyList{"enter", "o"}, keyContextArtifacts),
	keyDef(keyArtifactCopy, "y", KeyList{"y"}, keyContextArtifacts),

	keyDef(keyHelpClose, "esc", KeyList{"esc", "q", "?", "enter"}, keyContextHelp),
	keyDef(keyHelpDown, "j", KeyList{"j", "down"}, keyContextHelp),
	keyDef(keyHelpUp, "k", KeyList{"k", "up"}, keyContextHelp),
	keyDef(keyURLCopy, "y", KeyList{"y"}, keyContextURLList),
	keyDef(keyURLOpen, "o", KeyList{"o"}, keyContextURLList),
	keyDef(keyReactionClose, "esc", KeyList{"esc", "+", "a"}, keyContextReaction),
	keyDef(keyReactionLike, "1", KeyList{"1"}, keyContextReaction),
	keyDef(keyReactionHeart, "2", KeyList{"2"}, keyContextReaction),
	keyDef(keyReactionLaugh, "3", KeyList{"3"}, keyContextReaction),
	keyDef(keyReactionSurprised, "4", KeyList{"4"}, keyContextReaction),
	keyDef(keyReactionSad, "5", KeyList{"5"}, keyContextReaction),
	keyDef(keyReactionAngry, "6", KeyList{"6"}, keyContextReaction),
	keyDef(keyDeleteConfirm, "y", KeyList{"y", "Y"}, keyContextDeleteConfirm),
	keyDef(keyDeleteCancel, "n", KeyList{"n", "N", "esc"}, keyContextDeleteConfirm),
	keyDef(keyPresenceClose, "esc", KeyList{"esc", "q", "p", "enter"}, keyContextPresence),
	keyDef(keyPresencePrevious, "up", KeyList{"up", "k"}, keyContextPresence),
	keyDef(keyPresenceNext, "down", KeyList{"down", "j"}, keyContextPresence),
	keyDef(keyProfileClose, "esc", KeyList{"esc", "q", "I", "enter"}, keyContextProfile),
}

type resolvedKeyBinding struct {
	keyBindingDefinition
	Keys       KeyList
	Customized bool
}

// KeyMap is the runtime key dispatcher. It translates configured keys back to
// stable canonical keys, allowing the established handlers to stay mode-safe.
type KeyMap struct {
	byContext map[keyContext][]resolvedKeyBinding
	byAction  map[string]KeyList
}

func normalizeKeyName(value string) string {
	if value == " " {
		return "space"
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	switch lower {
	case "esc", "escape":
		return "esc"
	case "enter", "return":
		return "enter"
	case "space", "spacebar":
		return "space"
	case "tab":
		return "tab"
	case "up", "down", "left", "right", "home", "end":
		return lower
	case "pageup", "page-up":
		return "pgup"
	case "pagedown", "page-down":
		return "pgdown"
	}
	prefixes := []struct {
		long  string
		short string
		out   string
	}{
		{"control+", "c-", "ctrl+"},
		{"ctrl+", "", "ctrl+"},
		{"meta+", "m-", "alt+"},
		{"alt+", "", "alt+"},
		{"shift+", "s-", "shift+"},
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(lower, prefix.long) {
			rest := normalizeKeyName(value[len(prefix.long):])
			if len([]rune(rest)) == 1 {
				rest = strings.ToLower(rest)
			}
			return prefix.out + rest
		}
		if prefix.short != "" && strings.HasPrefix(lower, prefix.short) {
			rest := normalizeKeyName(value[len(prefix.short):])
			if len([]rune(rest)) == 1 {
				rest = strings.ToLower(rest)
			}
			return prefix.out + rest
		}
	}
	return value
}

func normalizeKeyList(values KeyList) KeyList {
	result := make(KeyList, 0, len(values))
	seen := make(map[string]bool)
	for _, value := range values {
		key := normalizeKeyName(value)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, key)
	}
	return result
}

func knownKeyActions() map[string]bool {
	known := make(map[string]bool)
	for _, definition := range keyBindingDefinitions {
		known[definition.Action] = true
	}
	return known
}

// DefaultKeyBindingConfig returns every built-in action and its default keys.
func DefaultKeyBindingConfig() KeyBindingConfig {
	result := make(KeyBindingConfig)
	for _, definition := range keyBindingDefinitions {
		if _, exists := result[definition.Action]; exists {
			continue
		}
		result[definition.Action] = append(KeyList(nil), definition.Defaults...)
	}
	return result
}

// NewKeyMap overlays user configuration over the built-in bindings. Unknown
// actions and same-mode conflicts are returned as warnings but do not prevent
// startup; the first listed action wins a conflict deterministically.
func NewKeyMap(overrides KeyBindingConfig) (KeyMap, []string) {
	resolved := DefaultKeyBindingConfig()
	known := knownKeyActions()
	warnings := make([]string, 0)
	for action, values := range overrides {
		if !known[action] {
			warnings = append(warnings, fmt.Sprintf("unknown keybinding action %q", action))
			continue
		}
		resolved[action] = normalizeKeyList(values)
	}

	keyMap := KeyMap{
		byContext: make(map[keyContext][]resolvedKeyBinding),
		byAction:  resolved,
	}
	for _, definition := range keyBindingDefinitions {
		_, customized := overrides[definition.Action]
		binding := resolvedKeyBinding{
			keyBindingDefinition: definition,
			Keys:                 append(KeyList(nil), resolved[definition.Action]...),
			Customized:           customized,
		}
		binding.Defaults = normalizeKeyList(binding.Defaults)
		for _, context := range definition.Contexts {
			keyMap.byContext[context] = append(keyMap.byContext[context], binding)
		}
	}
	for context := range keyMap.byContext {
		sort.SliceStable(keyMap.byContext[context], func(i, j int) bool {
			return keyMap.byContext[context][i].Customized && !keyMap.byContext[context][j].Customized
		})
	}

	contexts := make([]string, 0, len(keyMap.byContext))
	contextByName := make(map[string]keyContext)
	for context := range keyMap.byContext {
		name := string(context)
		contexts = append(contexts, name)
		contextByName[name] = context
	}
	sort.Strings(contexts)
	for _, contextName := range contexts {
		context := contextByName[contextName]
		owners := make(map[string]string)
		for _, binding := range keyMap.byContext[context] {
			for _, key := range binding.Keys {
				if prior, exists := owners[key]; exists && prior != binding.Action {
					warnings = append(warnings, fmt.Sprintf("key %q is assigned to both %s and %s in %s", key, prior, binding.Action, context))
					continue
				}
				owners[key] = binding.Action
			}
		}
	}
	sort.Strings(warnings)
	return keyMap, warnings
}

// Canonical returns the established handler key for a configured action. A
// removed default returns an empty string so overriding a binding really does
// disable the old key instead of letting it fall through to the legacy switch.
func (keyMap KeyMap) Canonical(context keyContext, pressed string) string {
	if keyMap.byContext == nil {
		return pressed
	}
	pressed = normalizeKeyName(pressed)
	bindings := keyMap.byContext[context]
	for _, binding := range bindings {
		for _, key := range binding.Keys {
			if key == pressed {
				return binding.Canonical
			}
		}
	}
	for _, binding := range bindings {
		for _, key := range binding.Defaults {
			if key == pressed {
				return ""
			}
		}
	}
	return pressed
}

func displayKeyName(key string) string {
	switch key {
	case "esc":
		return "Esc"
	case "enter":
		return "Enter"
	case "space":
		return "Space"
	case "tab":
		return "Tab"
	case "up":
		return "↑"
	case "down":
		return "↓"
	case "pgup":
		return "PgUp"
	case "pgdown":
		return "PgDown"
	}
	if strings.HasPrefix(key, "ctrl+") {
		return "Ctrl+" + strings.TrimPrefix(key, "ctrl+")
	}
	if strings.HasPrefix(key, "alt+") {
		return "M-" + strings.TrimPrefix(key, "alt+")
	}
	if strings.HasPrefix(key, "shift+") {
		return "Shift+" + strings.TrimPrefix(key, "shift+")
	}
	return key
}

// Display returns the active keys for help and popup labels.
func (keyMap KeyMap) Display(action string) string {
	keys, exists := keyMap.byAction[action]
	if !exists || len(keys) == 0 {
		return "unbound"
	}
	labels := make([]string, 0, len(keys))
	for _, key := range keys {
		labels = append(labels, displayKeyName(key))
	}
	return strings.Join(labels, " / ")
}

func (keyMap KeyMap) Primary(action string) string {
	keys := keyMap.byAction[action]
	if len(keys) == 0 {
		return "-"
	}
	return displayKeyName(keys[0])
}

func ResolveKeyMap() (KeyMap, []string) {
	cfg := LoadConfig()
	if cfg == nil {
		return NewKeyMap(nil)
	}
	return NewKeyMap(cfg.KeyBindings)
}
