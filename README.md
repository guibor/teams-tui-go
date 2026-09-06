# teams-tui-go

A standalone, keyboard-driven terminal client for Microsoft Teams.

It runs in any normal terminal on macOS, Linux, or Windows. No companion
editor, backend package, or private configuration repository is required.
Authentication uses the built-in OAuth2 device flow by default,
or an optional external short-lived token command when another application
already owns Microsoft sign-in.

This maintained fork builds on
[nospor/teams-tui-go](https://github.com/nospor/teams-tui-go) and adds a larger
conversation workflow, robust read-state navigation, complete exports,
bidirectional text rendering, configurable keybindings, component search,
new-chat creation, and external authentication support.

---

## Features

- 🔐 External token command or OAuth2 Device Code Flow — integrate with an existing credential owner without copying refresh tokens
- 💬 List all your Teams chats (1:1, group, meetings) with computed display names
- 📨 View messages in any chat with HTML-to-text rendering (images, attachments, emoji, **bold**, *italic*, ~~strikethrough~~, `code`, lists)
- 🛠️ Readable Teams system events — meeting/call starts and endings, durations, recordings, transcripts, membership changes, renames, and future Graph event types are shown as meaningful messages rather than a generic placeholder
- 🔤 Mixed Hebrew/English rendering for grid terminals, with ANSI styles and hyperlinks preserved and Hebrew-first lines aligned right
- ❤️ Message Interactions — view and add reactions (Heart, Like, Laugh, etc.) to any message
- 🔗 Clickable, Extractable & Openable URLs — links are clickable in supported terminals, can be extracted/copied via the `u` key, and opened in your browser/app via the `o` key
- ✏️ Message Management — send, edit, and delete messages (includes multi-line support)
- **✍️ Markdown Formatting** — compose messages with `**bold**`, `*italic*`, ~~`~~strike~~`~~, `` `code` ``, fenced code blocks, and bullet/ordered lists; formatting is sent as rich HTML to all Teams clients and rendered with ANSI styles in the TUI
- 📋 **Clipboard Image Pasting** — paste images from your system clipboard directly into the compose text field using **Ctrl+V** (automatically base64 encoded and sent as inline HTML attachments)
- 🗣️ **@Mentions & Autocomplete** — mention users in your messages. Typing `@` displays a dropdown list of chat/channel members. Navigate with Up/Down/Tab/Shift+Tab and press Enter to autocomplete the name. Mentions are sent as native Microsoft Teams mentions.
- 🔔 Notification modes: None / Console (BEL + visual bell) / System (desktop) / Both
- 🔄 Smart Background Polling & Sleep Mode — active chat messages poll every 3 s and chat list updates every 15 s. Polling auto-pauses when the terminal window is unfocused (blurred) or when you manually enter sleep mode via the `Esc` key.
- 😊 Emoticon Auto-replacement — popular text emoticons (like `:)`, `:D`, `<3`) are automatically converted to Unicode emojis
- 🔍 Search History — search messages in any chat, recursively loading and indexing all conversation history in the background
- 🔍 Chat and Message Search — literal/regexp components search a complete session inventory, with chat-name hits before participants and loaded message hits
- ➕ New Chats — press `N` to choose one or more participants, create/reuse a 1:1 or create a group, and open an empty composer without sending anything
- 🧭 Chat List Filters — press `v` or `V` to combine unread/read state, today's activity, chat type, favorites, and name/member text without making another Graph request
- 🔖 Chat Bookmarks — use quick two-key presets such as `bu` (unread), `bi` (inbox/all), `bt` (today), `bf` (favorites), `bd` (direct), `bg` (groups), and `bm` (meetings); `U` independently narrows any current view to unread
- ⭐ Favourites — pin any chat to the top of the sidebar with `*`; favourites are sorted alphabetically and stay anchored regardless of activity
- ↗️ Open in Teams — press `o` in normal mode to open the selected chat using Graph's native Teams URL and your configured browser/app command
- ❓ Help Popup — press `?` at any time to show a keyboard shortcuts reference with optional feature status

- 🔵 Read-State Styling — unread chats use a cyan dot with bright bold text; read chats are deliberately muted
- 📬 Explicit Read State — press `r` to mark a chat read or `u` to mark it unread (`i` remains a compatibility alias), then continue on the next visible chat; merely moving over a chat does not mark it read by default
- ✉️ Message Actions — use `c`/`C` to compose, uppercase `R` to reply with a quote, and `f`/`F` to forward an editable readable copy through the local chat chooser
- 📥 Full Markdown Export — press `E` to fetch every page of the selected chat and save a chronological Markdown transcript
- 🤖 External Thread Analysis — press `A` to make the same complete export and pass it to a configured analysis command
- ⌨️ Configurable Keys — override nearly every application action by stable name in `config.json`; defaults preserve the documented keyboard workflow and in-app help shows active bindings
- 😊 Reaction Indicators — chats with new reactions from other users are marked with their corresponding emoji (e.g. ❤️, 👍, 😆) and bold text
- ⬆️ New messages bubble chats to the top of the list
- 📌 Stable chat ordering — order only changes when new messages arrive
- 🧷 Selection-safe transcripts — filter and read-state refreshes reconcile the highlighted chat ID, transcript owner, and message conversation metadata; stale background responses cannot merge messages from different chats
- ↔️ RTL-safe compose — Hebrew and mixed-direction drafts use the same visual bidi renderer as messages while retaining logical Unicode order for editing and sending
- 🗓️ Clear conversation reading — day separators, explicit headers for each channel reply, concise conversation metadata, and one-for-one system events
- 🎥 Meeting resources — press `T` (or `a t`) to choose any loaded recording/transcript event, open its best available link, or copy the link
- 💾 Provider-aware refresh — external credentials stay with their owner; built-in device-flow tokens refresh from the application cache

**Optional features** (enable per-feature in `config.json`; see [AZURE_SETUP.md](AZURE_SETUP.md)):
- 📎 **File Preview & Download** (`file_preview_enabled`) — Tab through attachments and Teams-hosted inline images in the message popup and press Enter to download them to `~/Downloads/`
  - **Terminal Image Preview** (`file_preview_in_terminal`) — Displays the highlighted image inside the details popup using Kitty graphics or Sixel in compatible terminals (requires `file_preview_enabled: true`)
- ⬆️ **File Browsing & Uploading** (`file_upload_enabled`) — Press `Ctrl+f` in compose mode to open a file browser and attach small files (up to 50MB) from your computer. Files are uploaded to OneDrive/SharePoint and attached to your message.
- 🟢 **User Presence** (`presence_enabled`) — press `p` in message selection mode to see real-time availability of the message sender
- 👤 **User Profile** (`user_profile_enabled`) — press `I` in message selection mode to view extended profile info (name, email, job title, department)
- 🏢 **Teams Channels** (`teams_channels_enabled`) — Teams channels appear in the main sidebar below your chats; navigate with `j`/`k` and read messages just like chats. Supports background polling, global activity sorting (most active unhidden channels on top), unread indicators, and user-toggleable hidden channels (press `h` to toggle).

---

## Installation

### Install with Go

Go 1.25.4 or later is required when building from source:

```bash
go install github.com/guibor/teams-tui-go@main
teams-tui-go
```

The default build opens Microsoft device login on first use. Some corporate
tenants require administrator approval for third-party Graph clients; see
[Azure setup and permissions](AZURE_SETUP.md) when the default sign-in is
blocked or optional features need extra scopes.

### Release binary

When a tag includes prebuilt assets on
[GitHub Releases](https://github.com/guibor/teams-tui-go/releases), download the
archive for your operating system and architecture, extract `teams-tui-go`, and
put it on your `PATH`. Tagged source archives are always available from GitHub;
the Go installation above is the authoritative path when a tag has no binary
assets.

### Build from source

```bash
git clone https://github.com/guibor/teams-tui-go
cd teams-tui-go
go build -o teams-tui-go .

# or (builds slower, but binary is smaller)
go build -trimpath -ldflags="-s -w" -o teams-tui-go .

# then run
./teams-tui-go

# optionally install the binary on your PATH
sudo cp teams-tui-go /usr/local/bin/
```

---

## Configuration

Run this to locate the active JSON file:

```bash
teams-tui-go --config-path
```

Missing settings are added with backwards-compatible defaults at startup.

### Authentication

#### Built-in device login

No helper application is required. Run `teams-tui-go`, follow the displayed
Microsoft verification URL, enter the device code, and approve the requested
Graph permissions. The resulting refresh token is stored under the app's
private cache directory and refreshed by the TUI.

#### External token provider

Set `TEAMS_TUI_GO_TOKEN_COMMAND` to an executable path when another program
owns Microsoft authentication. The command takes no arguments and must print
one JSON object to stdout:

```json
{
  "access_token": "short-lived-graph-token",
  "expires_at": 1785686400
}
```

`expires_at` is a Unix timestamp in seconds. The TUI caches this access token
in memory until five minutes before expiry and then invokes the command again.
It never receives or stores the provider's refresh token.

When this variable is set, provider failure is fatal: the application does
**not** fall back to device-code authentication. The command must be a single
executable path, not a shell command with arguments.

Installers can verify support without starting authentication:

```bash
teams-tui-go --auth-provider-capabilities
# external-token-command-v1:auto
```

Packagers that must prohibit the built-in device flow can compile an
external-only binary. Such a binary rejects both new device login and any old
device-token cache when the external command is absent:

```bash
go build -ldflags="-X main.authMode=external-only" -o teams-tui-go .
```

#### Client ID (optional)

By default the app uses Microsoft's public Teams client ID. To use your own Azure AD app registration:

1. Follow the instructions in [AZURE_SETUP.md](AZURE_SETUP.md).
2. Set your client ID using one of:

   **Option A — environment variable:**
   ```bash
   cp .env.example .env
   # Edit .env and set CLIENT_ID=<your-client-id>
   ```

   **Option B — config file** (`~/.config/teams-tui-go/config.json`):
   ```json
   {
     "client_id": "your-client-id-here"
   }
   ```

### Keybindings

Nearly every application-level shortcut is configurable through the
`keybindings` object. Action names are stable and mode-qualified, so the same
physical key can safely mean different things in the chat list, message view,
and composer.

```json
{
  "keybindings": {
    "search.global": ["ctrl+s"],
    "new_chat.open": "ctrl+n",
    "filter.toggle_unread_overlay": ["U", "ctrl+u"],
    "compose.send_prefix": "ctrl+c",
    "compose.send": ["ctrl+s", "ctrl+enter"],
    "chat.analyze": []
  }
}
```

A string and an array are both accepted. Mentioning an action replaces its
defaults; `[]` deliberately unbinds it. Omitted actions retain their built-in
keys. Common key spellings such as `C-s` and `M-n` are accepted as aliases
for `ctrl+s` and `alt+n`. Restart the TUI after editing the file.

Print the complete authoritative action map with:

```bash
teams-tui-go --print-default-keybindings
```

The `?` popup displays the active bindings. Duplicate keys assigned to two
actions in the same mode produce a startup warning. Text editing remains under
the terminal text widget, the file browser retains its own navigation map, and
bookmark suffixes (`bu`, `bt`, and custom presets) remain configured through
`chat_bookmarks`.

### Notifications
- **Toggle Mode**: Cycle through notification modes at runtime by pressing `n`. The chosen mode is automatically saved.
- **Message Previews**: Configure desktop notifications in `~/.config/teams-tui-go/config.json`:

  ```json
  {
    "notification_mode": "System",
    "notification_show_preview": true,
    "notification_preview_len": 80
  }
  ```
  - `notification_show_preview`: Set to `true` to include the message content in the desktop notification.
  - `notification_preview_len`: The maximum number of characters to show in the preview.

### Message Limit
Configure how many messages to fetch when opening a chat in `~/.config/teams-tui-go/config.json`:

  ```json
  {
    "message_limit": 50
  }
  ```
  - `message_limit`: The number of messages to fetch (default: 50). For limits greater than 50, the app automatically makes sequential paginated requests. Capped at `200` to prevent excessive API requests.

### Chat Limit
Configure how many chats to load in the sidebar in `~/.config/teams-tui-go/config.json`:

  ```json
  {
    "chat_limit": 50
  }
  ```
  - `chat_limit`: The maximum number of chats to fetch and display (default: 50). Automatically makes paginated requests if needed. Capped at `200` to bound member-loading work.

### Read State and Markdown Exports

Navigation is non-destructive by default. Select a chat and press `r` to mark
it read (`i` is also accepted) or `u` to mark it unread. A successful dispatch
immediately selects and loads the next visible chat, wrapping at the end. To
restore the original mark-on-open behavior, set `mark_read_on_open` to `true`.
Press `C-r` to force an immediate chat/read-state refresh. Its completion
status reports Graph viewpoint coverage and the selected chat's authoritative
Graph state, making client discrepancies diagnosable.

Filtering and background refreshes preserve selection by chat ID. The next
chat is captured by ID before a read-state action can change the visible list,
so a filtered-list rebuild cannot skip another row. Non-disposition actions,
including export, capture, analysis, open, favorite, and recording/transcript
actions, retain the current chat.

Each chat refresh also reconciles Microsoft Graph's per-user `viewpoint`.
The initial server timestamp establishes read state; a later timestamp change
reflects reads or unread marks made in Teams or another client. Unchanged
server timestamps do not roll back a newer local action while Graph converges.
The comparison is exact, preserving the sub-second read-pointer movement used
when Teams marks only the latest message unread.
Messages from the
previous chat are never retained or merged into the replacement pane.

The sidebar has one fixed summary header. Applying a filter/bookmark or toggling
the date column requests a clean terminal repaint so a stale header cannot
remain above the current one. The selected chat or channel is rendered as a
continuous full-width, bold white-on-blue row; marker and icon ANSI resets
cannot cancel that highlight partway across the line.

`M-n` and `M-p` select the next and previous chat in the visible list. They use
the same filtered-list navigation as `j` and `k`; `M-<` and `M->` jump to the
first and last visible item in the active chat or channel section. In the chat
list, `gg` or `h` selects the first visible chat and `G` or `l` selects the last.
Plain `<`/`>` and `H`/`L` jump to the top/bottom of the loaded message pane.

Teams-generated system messages use Microsoft Graph's `eventDetail` metadata.
The normal thread view, history search, notifications, message popup, and
Markdown exports therefore show text such as `Meeting started`,
`Meeting ended (23m 15s)`, `Call recording available`, or
`Call transcript available`. Unknown future event types are humanized instead
of being hidden behind a generic system-event label.

The conversation pane does not collapse duplicate system events. Press `T`, or
choose `t` from thread actions, to list each loaded recording and transcript
event separately. Recording events prefer their direct recording URL;
transcripts open the Teams event/message link because Graph does not expose a
direct transcript URL in the event detail.

Press `D` to toggle a fixed-width local last-message date and 24-hour time
immediately before each chat title. Current-year rows use `Aug 03 17:45`; older
rows include the year as in `2025-12-31 17:45`. The choice persists as
`show_chat_dates` in `config.json`.

Chat and channel names in the sidebar always occupy one row and end in an
ellipsis when the available terminal-cell width is exhausted, including while
selected. The complete selected conversation name appears at the top of the
wider message pane.

Press `E` to fetch the complete selected chat, follow every Graph pagination
link, and write a chronological Markdown transcript. Exports default to
`~/Downloads`; change `export_directory` to use another location. Press `a`
then `a` to add the selected chat under today's heading in a deduplicated
thread list. Markdown checkboxes are the default. Set `thread_capture_format`
to `org` to write an Org `TODO` capture with a timestamp, Teams properties,
and source link to `thread_capture_org_file` instead.

  ```json
  {
	"mark_read_on_open": false,
	"default_snooze_minutes": 180,
	"workday_start": "07:00",
	"workday_end": "18:00",
	"show_chat_dates": false,
	"export_directory": "~/Downloads",
	"thread_capture_format": "markdown",
	"thread_capture_file": "~/Documents/teams-threads.md",
	"thread_capture_org_file": "~/Documents/teams-threads.org"
  }
  ```

`thread_capture_format` accepts `markdown` or `org` and is read at startup.
The two destination settings are kept separately, so switching formats does
not mix Org syntax into the Markdown list or discard either file.

Press `A` (or `a A` from thread actions) to fetch and save that same complete
transcript in the background, then invoke `thread_analysis_command` with
`--agent AGENT PATH`. The command is intentionally unconfigured by default;
set it to a local bridge that accepts those arguments and starts the desired
analysis tool. A bridge can, for example, submit a first prompt such as:

```text
$thread-analysis of this thread: /absolute/path/to/export.md
```

Set `thread_analysis_agent` to `codex`, `agent`/`cursor`, `claude`/
`claude-code`, `cline`, `goose`, `default`, or another identifier understood by
your analysis command. Export completes before the command runs; if launch fails, the
status line retains the saved Markdown path for manual recovery.

```json
{
  "export_directory": "~/Downloads",
  "thread_analysis_agent": "codex",
	"thread_analysis_destination": "terminal",
	"thread_analysis_model": "gpt-5.6-luna",
	"thread_analysis_models": ["gpt-5.6-luna", "gpt-5.6-sol", "gpt-5.6-terra", "default"],
  "thread_analysis_command": "/usr/local/bin/thread-analysis-bridge"
}
```

The configured action (`A`) shows its command, destination, agent, and model in
the thread-actions window. The interactive action (`X`) asks for a destination
(`terminal`, `emacs`, or `codex-app`) and then a model from
`thread_analysis_models`. Destination and model are exposed to the bridge as
`TEAMS_THREAD_ANALYSIS_DESTINATION` and `TEAMS_THREAD_ANALYSIS_MODEL`; the
existing `--agent AGENT PATH` command contract remains unchanged. Commands may
also contain `{destination}`, `{model}`, and `{agent}` placeholders when the
selection must be passed explicitly as command-line arguments.

### Search Queries

The global `s` chooser loads every paginated chat once per session into a
transient inventory, including chats outside `chat_limit` or hidden by the
active sidebar view. Typing remains entirely local. The inventory is not added
to the sidebar or SQLite; only a chat that you open is hydrated into the normal
chat model. Results are sectioned in this order: chat title/topic, participant
name/email, loaded messages/latest previews, then channels. The
current-conversation `/` history search uses the same query grammar while its
existing recursive history loader continues to fetch older pages.

Each space-separated component may match literally or as a regexp, every
component is required, and
component order does not matter. There is no arbitrary character-subsequence
matching: `q p` and `q.*p` match “Quarterly Planning”; `qp` does not. Quote a
phrase, prefix a term with `-` to exclude it, and use these fields:

| Syntax | Meaning |
| --- | --- |
| `from:alice` | Sender name |
| `in:"Product planning"` | Conversation name |
| `is:unread`, `is:read`, `is:favorite` | Chat state |
| `type:direct`, `type:group`, `type:meeting` | Chat type |
| `type:message`, `type:event`, `type:channel` | Result type |
| `has:file`, `has:image`, `has:link` | Message/latest-preview content |
| `after:2026-08-01`, `before:2026-08-06` | Local message/activity date |

For example, `quarter plan from:alice is:unread -has:file` matches both free
components in any order in unread conversations while excluding file-bearing
messages.

### New Chats

Press `N` to open the participant picker. Known participants from the transient
chat inventory appear immediately; after a short pause, tenant-directory
results are added when the token has `User.ReadBasic.All` or `User.Read.All`.
An exact email/UPN remains available when directory search is unavailable.

Use `Enter` on the input's first match, or move into the result list and use
`Space`/`Enter`, to add or remove participants. `Ctrl+Enter`/`Ctrl+J` creates or
reuses a 1:1 for one participant and creates a group for multiple participants.
The resulting chat opens in an empty composer and is not sent until the normal
compose send shortcut is used.

### Search Context Limit
Configure how many context messages (before and after each search match) to display in the search history popup in `~/.config/teams-tui-go/config.json`:

  ```json
  {
    "search_context_limit": 3
  }
  ```
  - `search_context_limit`: The number of context messages before and after each match to include (default: 3).

### Channel Message Refresh
Configure background refresh rate for unhidden channels (in minutes) in `~/.config/teams-tui-go/config.json`:

  ```json
  {
    "channel_msg_refresh_min": 2
  }
  ```
  - `channel_msg_refresh_min`: The background polling interval in minutes for unhidden channels (default: 2).

### External Editor
Configure the external editor to open when pressing `Ctrl+g` in compose mode in `~/.config/teams-tui-go/config.json`:

  ```json
  {
    "external_editor": "emacsclient --wait"
  }
  ```
  - `external_editor`: The editor command and optional arguments to run (e.g. `"vim"`, `"nvim -f"`, or `"emacsclient --wait"`). The temporary message path is appended as the final argument. If empty/unspecified, it falls back to `$EDITOR`, then `$VISUAL`, and defaults to `"vim"`.

### URL Opening Commands
Configure the commands used to open URLs when pressing `o` on a message or from the URL selection menu in `~/.config/teams-tui-go/config.json`:

  ```json
  {
	"browser_command": "xdg-open",
	"teams_app_command": "xdg-open",
    "youtrack_command": "yt-tui",
    "gitlab_command": "gitlab-tui"
  }
  ```
	- `browser_command`: The command used to open general URLs (default: `"xdg-open"`, but you can specify e.g. `"firefox"` or `"google-chrome"`). This key is always initialized in `config.json`.
	- `teams_app_command`: The command used by uppercase `O` to dispatch the selected chat's `msteams://` deep link to the installed Teams desktop client. It defaults to `open` on macOS, `xdg-open` on Linux, and the Windows URL protocol handler on Windows.
  - `youtrack_command`: The optional command to open YouTrack URLs (default: `"yt-tui"`, but you can specify e.g. `"youtrack-cli"` or `"yt-cli"`). If a URL contains `"youtrack"`, this command is executed. Useful with tools like [yt-tui](https://github.com/nospor/yt-tui).
  - `gitlab_command`: The optional command to open GitLab URLs (default: `"gitlab-tui"`). If a URL contains `"gitlab"` (for example, merge requests, pipelines, or jobs), this command is executed. Useful with tools like [gitlab-tui](https://github.com/nospor/gitlab-tui).

### Image Viewer
Configure a dedicated image viewer used when pressing `Enter` on an image attachment (inside the `v` popup with `Tab` to enter attachment cursor mode).
When set, **all** image attachments from the current message are downloaded and loaded into the viewer together, automatically starting at the selected one.

  ```json
  {
    "image_viewer": "sxiv"
  }
  ```
  - `image_viewer`: The command or executable to open image attachments (e.g. `"sxiv"`, `"nsxiv"`, `"feh"`, `"imv"`, or any viewer). If empty or not set (the default), images fall back to the regular single-file opener (`xdg-open` / `open`). When set, pressing `Enter` on an image attachment downloads **all** image attachments in that message and loads them all into the viewer, automatically starting at the selected one. Viewer-specific start-index flags are handled automatically:
    - **`sxiv`** / **`nsxiv`**: uses `-n INDEX` (1-based).
    - **`feh`**: uses `--start-at PATH`.
    - **`imv`**: uses `-n PATH`.
    - **Other viewers**: the selected image is passed first in the argument list.

### Chat Icon Themes
You can configure the style of chat type indicators in the sidebar using `~/.config/teams-tui-go/config.json`:

```json
{
  "chat_icon_theme": "unicode"
}
```
- `chat_icon_theme`: Choose between presets (default: `"unicode"`):
  - `"unicode"`: Clear single-width symbols (`@` 1:1, `&` group, `◷` meeting, `#` channel).
  - `"emoji"`: Colorful emojis (`👤` 1:1, `👥` group, `📅` meeting, `#️⃣` channel).
  - `"text"`: The original bracketed text headers (`[oneOnOne]`, `[group]`, `[meeting]`, `[channel]`).

You can also completely override icons individually by defining a `"custom_chat_icons"` map:

```json
{
  "custom_chat_icons": {
    "oneOnOne": "💬",
    "group": "👥",
    "meeting": "⏱️",
    "channel": "📢",
    "default": "◈"
  }
}
```

### Optional Features

Each feature is disabled by default and requires an additional Graph API permission. Enable them in `~/.config/teams-tui-go/config.json` and **delete `~/.cache/teams-tui-go/token.json`** to force re-authentication with the new scopes:

On macOS, direct and wrapped launches both use this portable `~/.config` path
rather than `~/Library/Application Support`. Run `teams-tui-go --config-path`
or open `?` help to confirm the effective file, then restart the TUI after edits.

```json
{
  "sqlite_enabled": false,
  "file_preview_enabled": true,
  "file_preview_in_terminal": false,
  "terminal_image_protocol": "auto",
  "file_upload_enabled": false,
  "presence_enabled": true,
  "user_profile_enabled": true,
  "user_profile_extended": false,
  "teams_channels_enabled": false,
  "channel_mentions_enabled": false
}
```

| Config key | Default | Required permission | Effect |
|---|---|---|---|
| `sqlite_enabled` | `false` | - | Enables offline caching via SQLite (`~/.cache/teams-tui-go/teams-tui-go.db`). Instantly loads messages when entering chats/channels, syncing updates in the background. |
| `file_preview_enabled` | `false` | `Files.Read` | Tab through attachments in the `v` popup and press Enter to download to `~/Downloads/` |
| `file_preview_in_terminal` | `false` | `Files.Read` | Previews the highlighted file attachment or Teams-hosted inline image inside the details popup (requires `file_preview_enabled: true`) |
| `terminal_image_protocol` | `auto` | - | Chooses `sixel` inside Emacs EAT and `kitty` elsewhere; accepts explicit `kitty`, `sixel`, or `none`, and `TEAMS_TUI_GO_IMAGE_PROTOCOL` overrides it |
| `file_upload_enabled` | `false` | `Files.ReadWrite` | Press `Ctrl+f` in compose mode to open a file browser and attach files under 4MB from the computer |
| `presence_enabled` | `false` | `Presence.Read.All` | Press `p` in message selection mode to see sender availability |
| `user_profile_enabled` | `false` | `User.ReadBasic.All` | View sender profiles and add tenant-directory matches to the `N` participant picker |
| `user_profile_extended` | `false` | `User.Read.All` *(admin consent)* | Adds job title, department, office to the profile popup (requires `user_profile_enabled: true`) |
| `teams_channels_enabled` | `false` | `Team.ReadBasic.All` + `Channel.ReadBasic.All` + `ChannelMessage.Read.All` *(admin consent)* + `ChannelMessage.Send` + `ChannelMessage.ReadWrite` | Teams channels appear in the sidebar below chats; navigate with `j`/`k`. Supports background polling, activity sorting, unread dots, and hidden channels (`h` key). |
| `channel_mentions_enabled` | `false` | `TeamMember.Read.All` | Enables autocomplete suggestion dropdown list of team members in Teams channels when typing `@` mentions. |

When `TEAMS_TUI_GO_TOKEN_COMMAND` is configured, the external provider owns
authentication and the built-in `token.json` cache is not used. Feature flags
still expose the corresponding UI, but the provider's Graph token must already
carry the listed permission.

### Bidirectional Text

Most grid terminals place Unicode code points strictly from left to right. The
message, search, and detail views therefore apply Unicode bidirectional
reordering after logical line wrapping. An English-first mixed line remains
left-to-right while its Hebrew runs display correctly; a Hebrew-first line is
also aligned to the right edge. ANSI styling, OSC 8 links, emoji sequences, and
combining marks remain attached to their original grapheme clusters. Copy,
edit, and Markdown export paths retain logical text order.

See [AZURE_SETUP.md](AZURE_SETUP.md) for full permission setup instructions.

---

## Usage

```bash
# Run directly
./teams-tui-go

# Or if installed
teams-tui-go
```

Without `TEAMS_TUI_GO_TOKEN_COMMAND`, first run prompts you to visit a URL and
enter a short code; subsequent runs use the cached device-flow token. With an
external provider, authentication and refresh behavior are owned entirely by
that provider.

---

## Markdown Formatting

Messages support a subset of markdown that is converted to rich HTML when sent, so recipients on **any Teams client** (Desktop, Web, Mobile) see proper formatting.

### Inline syntax

| Syntax                   | Result     |
| ------------------------ | ---------- |
| `**bold**` or `__bold__` | **Bold**   |
| `*italic*` or `_italic_` | *Italic*   |
| `~~strikethrough~~`      | ~~Strike~~ |
| `` `inline code` ``      | `code`     |

### Block syntax

| Syntax                        | Result                  |
| ----------------------------- | ----------------------- |
| `* item` or `- item`          | Unordered (bullet) list |
| `1. item` or `1) item`        | Ordered (numbered) list |
| ` ``` ` fence on its own line | Multi-line code block   |

**Example:**

````
**Meeting notes** for *Project X*

* Review PR #42
* Deploy to staging

```
fmt.Println("hello")
```
````

> **Note:** Language hints (e.g. ` ```go `) are accepted syntax but have no visible effect — Teams strips the `class` attribute from the stored HTML, so the hint is not preserved or displayed.

### Receive side rendering

Incoming messages from all clients are rendered with matching ANSI styles in the TUI:

- Bold, italic, and strikethrough use terminal text attributes
- Inline `code` is highlighted in amber
- Code blocks are highlighted in green
- Bullet and numbered lists show `•` / `1.` prefixes

### Edit round-trip

When you press `e` to edit an existing message the edit box is pre-filled with the **original markdown source** (e.g. `**bold**` rather than stripped plain text), so formatting is preserved after saving.

### Clipboard Image Pasting

When in compose mode (`c` or `C`), you can paste images (PNG/JPEG) directly from your system clipboard using **`Ctrl+V`**.
- A placeholder like `[Image 1]` will be inserted into the text field.
- You can move, copy, or delete this placeholder to control where the image appears in the sent message. If deleted, the image won't be sent.
- When the message is sent, the image is automatically base64-encoded and uploaded inline.

> [!NOTE]
> On Linux, `Ctrl+Shift+V` is intercepted by most terminal emulators to perform text-only paste. To paste clipboard images, make sure to use **`Ctrl+V`** instead, which is passed directly to the TUI.

### File Browsing & Uploading

When `file_upload_enabled` is set to `true` in `config.json`, you can attach small files (under 4MB) from your local computer to chat or channel messages.
- In compose mode (`c` or `C`), press **`Ctrl+f`** to open the offline file browser overlay.
- Navigate directories using `j`/`k` (or arrow keys) and enter directories with `Enter`. Move to parent directories via `..`.
- Highlight a file and press **`Enter`** to select and attach it.
- A placeholder like `[File: filename.ext]` is inserted into the textarea. You can move, copy, or delete it to control inline message rendering.
- When sending the message, files are automatically uploaded to OneDrive (for chats) or SharePoint (for channels) and attached as reference attachments to the message.

### External Editor (Composing & Viewing)

You can use an external editor (such as `vim`, `neovim`, or `nano`) to either compose a new message or view an existing message:

- **Composing/Editing**: When in compose mode (`c` or `C`), press **`Ctrl+g`** to open the external editor. The current input text is saved to a temporary file and loaded into the editor. When you save and exit, the TUI loads the changes back into the compose field.
- **Viewing**: When in message mode (`m`) or message details popup (`v`), press **`Ctrl+g`** to open the selected message's full content in the external editor in read-only mode. Edits made in the editor are discarded when you exit, returning you directly to your previous position in the TUI.

The external editor command can be configured in your `config.json` via the `"external_editor"` option. If not specified, it falls back to the `$EDITOR` environment variable, then `$VISUAL` environment variable, and defaults to `"vim"`.

Compose mode is multiline by default: `Enter` inserts a line break and
`Ctrl+Enter` sends the message. `Ctrl+J` is a portable send alias for terminals
that cannot distinguish modified Return. The TUI also recognizes the common
Kitty/iTerm `CSI 13;5u` Ctrl+Enter encoding. Emacs-style `Ctrl+C Ctrl+C` also
sends; its prefix is configurable with `compose.send_prefix`.

The forward destination chooser starts with known chats and uses the same
literal/regexp components. `q p` can match `Quarterly Planning`, while
the looser `qp` abbreviation does not. `Enter` accepts the highest-ranked local
match directly. An exact email/UPN creates or opens a direct chat only when no
local destination matches.

---

## Keyboard Controls

These are the default bindings. The in-app `?` view always shows the active
configuration; use the `keybindings` object to replace any application action.

| Key          | Action                                                    |
| ------------ | --------------------------------------------------------- |
| `↑` / `k`    | Move up in list (within active section)                   |
| `↓` / `j`    | Move down in list (within active section)                 |
| `M-p` / `M-n`| Move to previous / next item in the active section        |
| `M-<` / `M->`| Jump to first / last item in the active section            |
| `gg` / `G`   | Jump to first / last visible chat                          |
| `h` / `l`    | Jump to first / last visible chat                          |
| `<` / `>`    | Jump to top / bottom of the loaded messages pane           |
| `H` / `L`    | Jump to top / bottom of the loaded messages pane           |
| `Tab`        | Switch between Chats & Channels sections (in Normal Mode) |
| `PgUp` / `K` | Scroll messages up                                        |
| `PgDn` / `J` | Scroll messages down                                      |
| `/`          | Open search input (in Normal Mode)                        |
| `Esc`        | Clear active search, or leave conversation for dashboard   |
| `s`          | Component-search all chats and loaded messages             |
| `v` / `V`    | Filter chats by read state, type, favorite, and text      |
| `U`          | Toggle unread-only over the active bookmark/filter         |
| `b`          | Open chat bookmarks (`bu` unread, `bt` today, `b2` 24h)   |
| `z` / `Z`    | Snooze for the default duration / choose a wake time       |
| `a`          | Open actions for the selected chat                         |
| `T`          | Choose a loaded recording or transcript                    |
| `*`          | Toggle ★ favourite on selected chat (chats only)          |
| `o`          | Open selected chat directly in Teams web in the configured browser |
| `O`          | Open selected chat in the Teams desktop client            |
| `M`          | Join a meeting, or call the participant in a direct chat   |
| `r` / `i`    | Mark selected chat read, then advance to the next chat     |
| `u`          | Mark selected chat unread, then advance (Normal Mode)     |
| `c` / `C`    | Compose a new message in the current conversation          |
| `N`          | Choose participants and create a new 1:1 or group chat      |
| `R`          | Reply to the newest loaded message                         |
| `f` / `F`    | Forward the newest loaded message through the chat chooser |
| `E`          | Export complete selected chat as Markdown (Normal Mode)   |
| `A`          | Export complete chat and run configured analysis          |
| `h`          | Toggle hide/unhide on selected channel (channel mode)     |
| `Ctrl+V`     | Paste image from clipboard (in Compose Mode)              |
| `Ctrl+f`     | Browse and attach file from computer (in Compose Mode)    |
| `Ctrl+g`     | Compose/edit message in external editor (in Compose Mode) |
| `Ctrl+Enter` / `Ctrl+J` / `Ctrl+C Ctrl+C` | Send message                 |
| `Enter`      | Insert a normal new line                                  |
| `Alt+Enter` / `Shift+Enter` | Insert a new line                             |
| `Esc`        | Cancel compose                                            |
| `n`          | Toggle notification mode                                  |
| `?`          | Show help popup (keyboard reference + feature status)     |
| `m`          | Enter/Exit **Message Mode** (to select/react/delete/copy) |
| `<` / `>` or `H` / `L` | Select oldest / newest loaded message       |
| `v`          | View details/reactions of selected message (Message Mode) |
| `Ctrl+g`     | View selected message in external editor (in Message Mode / Message View Popup) |
| `Tab`        | Switch to attachment cursor in `v` popup (in Message View Popup) |
| `Enter`      | Download selected attachment (in `v` attachment cursor)   |
| `+` / `a`    | React to selected message (in Message Mode)               |
| `c` / `C`    | Compose without quoting (Message Mode / Message View)     |
| `r` / `i`    | Mark the selected message's conversation read and advance |
| `R`          | Reply to selected message (Message Mode / Message View)   |
| `f` / `F`    | Forward selected message (Message Mode / Message View)    |
| `y`          | Copy (yank) message text (in Message Mode)                |
| `u`          | Copy (yank) URL from message (in Message Mode / History Search) |
| `o`          | Open URL from message (in Message Mode / History Search / URL list) |
| `g`          | Go to/jump to message in normal view (in History Search results) |
| `d`          | Delete selected message (in Message Mode)                 |
| `e`          | Edit selected message (in Message Mode)                   |
| `p`          | Show presence status of sender (`presence_enabled`)       |
| `I`          | Show profile info of sender (`user_profile_enabled`)      |
| `1-6`        | Send reaction (in Reaction Mode)                          |
| `q`          | Leave conversation for dashboard; press again to quit     |
| `Ctrl+C` / `Q` | Quit immediately                                         |

### Chat List Filter

Press `v` or `V` from normal mode. Filters are local and combine with AND semantics, so
an unread + group + favorites filter shows only chats matching all three.
Press `U` from normal mode to toggle an unread predicate over the current filter
or bookmark. For example, `bt` then `U` means Today AND Unread; another `U`
returns to Today. The overlay survives other bookmark changes. `bi`/`ba`
explicitly reset it, while `bu` remains the standalone Unread bookmark.

| Filter key       | Action                                             |
| ---------------- | -------------------------------------------------- |
| `u` / `r` / `a`  | Unread only / read only / all read states          |
| `t`              | Toggle activity-today-only                         |
| `1` / `g` / `m`  | Toggle 1:1 / group / meeting chat types            |
| `f`              | Toggle favorites-only                              |
| `/`              | Edit component/structured query                    |
| `Space`          | Toggle or cycle the selected filter row            |
| `x`              | Clear the draft filter                             |
| `Enter`          | Apply                                               |
| `Esc`            | Cancel and retain the currently active filter      |

### Chat Bookmarks

Press `b` from normal mode and then a preset key. The popup also supports
`j`/`k` and `Enter`. Presets replace the active filter and remain entirely
local; applying one does not issue a Graph list request.

| Bookmark | Chats shown                         |
| -------- | ----------------------------------- |
| `bu`     | Unread                              |
| `br`     | Read                                |
| `bi`     | Inbox/all chats (clears filters)    |
| `ba`     | All chats (alias for `bi`)          |
| `bt`     | Activity on the current local day   |
| `b2`     | Activity in the rolling last 24 hours |
| `bw`     | Activity in the rolling last 7 days |
| `bs`     | Locally snoozed chats and their wake state |
| `bf`     | Favorites                           |
| `bd`     | Direct (1:1)                        |
| `bg`     | Groups                              |
| `bm`     | Meetings                            |

Add custom bookmarks with `chat_bookmarks` in `config.json`. A custom bookmark
with the same one-character key replaces that built-in entry; other keys are
appended to the popup.

```json
{
  "chat_bookmarks": [
    {
      "key": "p",
      "name": "Product planning",
      "query": "in:\"Product planning\"",
      "chat_types": ["group"]
    },
    {
      "key": "u",
      "name": "Urgent unread",
      "query": "urgent",
      "read_state": "unread",
      "favourites_only": true
    }
  ]
}
```

Bookmark records support `query`, `read_state`, `chat_types`,
`favourites_only`, `today_only`, and `within_hours`. `within_hours` is a
rolling window, unlike `today_only`, which follows the local calendar day.

`U` can be toggled after any bookmark and is ANDed with that bookmark. The
sidebar header shows both states, for example `Today · Unread`.

### Snooze

`z` hides the selected chat for `default_snooze_minutes` (three hours by
default) and advances by stable chat identity. `Z` offers ten minutes, one
hour, three hours, end of workday, tomorrow morning, next week, and unsnooze. Tomorrow
uses `workday_start` (07:00 by default); end of workday uses `workday_end`
(18:00 by default), falling back to the next workday morning after that time.
Snoozes are local, survive restarts, never alter Teams read state, and wake
immediately when a new incoming message arrives.

### Thread Actions

Press `a` on a selected chat. Use the action's displayed shortcut directly, or
navigate with `j`/`k` and run it with `Enter`.

| Action key | Action                                      |
| ---------- | ------------------------------------------- |
| `o` / `O`  | Open directly in Teams web / Teams desktop  |
| `c` / `C`  | Compose a message                           |
| `R`        | Reply to the latest message                 |
| `f` / `F`  | Forward the latest message                 |
| `r` / `u`  | Mark read / unread (`i` also marks read)    |
| `*`        | Toggle favorite                             |
| `a`        | Capture in the configured dated thread list (`aa`) |
| `e`        | Export the complete Markdown transcript     |
| `A`        | Export and analyze the complete thread (`a A`) |
| `y`        | Copy the Teams web link                     |
| `t`        | Choose a recording or transcript            |

Completed actions advance to the next visible chat, wrapping at the end. The
next chat is chosen before read/favorite filters can remove or reorder the
current row. Compose, reply, forward, and the recording/transcript chooser stay
on the current chat until their interactive workflow is complete.

---

## File Locations

| File                                          | Purpose                             |
| --------------------------------------------- | ----------------------------------- |
| `~/.config/teams-tui-go/config.json`           | Authentication, features, limits, bookmarks |
| `~/.config/teams-tui-go/favourites.json`       | Pinned/favourite chat IDs           |
| `~/.cache/teams-tui-go/token.json`             | Built-in device-flow tokens; unused with an external provider |
| `~/.cache/teams-tui-go/profile.json`           | Cached user profile                 |
| `~/.cache/teams-tui-go/teams-tui-go.db`       | SQLite database caching messages    |
| `~/Downloads/*.md`                            | Complete chat exports by default    |
| `~/Documents/teams-threads.md`                 | Dated thread-capture checklist by default |
| `~/Documents/teams-threads.org`                | Dated Org TODO captures when configured |

---

## Development

```bash
# Run in development
go run .

# Build binary
go build -o teams-tui-go .

# Lint
go vet ./...
```

---

## License

See [LICENSE](LICENSE).

## Acknowledgements

This fork retains the history and substantial foundation of
[nospor/teams-tui-go](https://github.com/nospor/teams-tui-go). Thanks to the
original maintainer and contributors who built the client this work extends.
