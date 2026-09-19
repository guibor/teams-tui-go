# Fork Changes And Upstream Status

This repository is an independent MIT-licensed fork of
[nospor/teams-tui-go](https://github.com/nospor/teams-tui-go). The original
copyright notice, license, and Git history are retained. Please file issues
about this fork in [this repository](https://github.com/guibor/teams-tui-go/issues).
Neither project is an official Microsoft client.

## Foundation And Additions

The original client provides the Go/Bubble Tea application, Graph integration,
device login, chats and optional channels, message composition and formatting,
reactions, mentions, attachments, external editing, notifications, favourites,
history search, and optional SQLite history.

The fork's work since the shared base includes:

| Area | Fork additions or extensions |
| --- | --- |
| Inbox workflow | Explicit read/unread, configurable bookmarks, rolling activity windows, unread overlay, local persistent snoozes, manual refresh |
| Navigation | Selection by chat identity across filters and async responses, transcript ownership checks, long-title truncation, dashboard-first quit |
| Search | Literal/regexp components in any order, title/member/message result sections, all-chat session inventory, new-chat participant picker |
| Messages | Editable forward copies, quoted reply workflow, multiline/send controls, clearer system events, day separators, mixed-direction rendering and compose |
| Export and capture | Complete paginated Markdown exports, dated Markdown/Org capture lists, optional external analysis handoff |
| Integrations | Separate web/desktop links, meeting/call links, recording/transcript chooser, additional terminal image protocols |
| Configuration | Mode-aware action keybindings, portable config location, analysis routes/models, external token and read-state provider interfaces |

These are differences from the shared base, not a claim that upstream's current
release lacks every comparable capability. Optional analysis is an executable
bridge contract; users supply the bridge and model/route identifiers. No
particular editor or terminal emulator is required.

## Upstream Snapshot: 2026-09-19

The shared ancestor is
[`6187c59`](https://github.com/nospor/teams-tui-go/commit/6187c592b5142182e223bf2b8fff643ae67c75e9),
dated 2026-07-24 (v1.2.3 changelog). The first fork-specific commit is dated
2026-08-02.

Upstream's fetched `main` was `969f60b`, dated 2026-09-11, with the v1.2.9
changelog. Before this publication cleanup, fork `7d671cf` had 52 unique commits
and upstream had 17 unique commits since their shared ancestor. The upstream
count includes changelog and merge commits.

All 17 upstream commits through `969f60b` have now been merged with their
history preserved. The integration keeps external-only authentication,
configurable key dispatch, transcript ownership checks, RTL rendering, and
readable system events. It also tests snooze expiry with cached rendering and
uses temporary files so failed downloads do not poison the attachment cache.

Notable upstream work included in the merge:

| Commit | Change |
| --- | --- |
| [64e31bb](https://github.com/nospor/teams-tui-go/commit/64e31bb) | Reduce idle CPU usage from the heartbeat/render loop |
| [35627cd](https://github.com/nospor/teams-tui-go/commit/35627cd) | Fix Enter handling when composing an email address |
| [ab1ffdf](https://github.com/nospor/teams-tui-go/commit/ab1ffdf) | Toggle hidden files/folders in the file picker |
| [08c89f0](https://github.com/nospor/teams-tui-go/commit/08c89f0) | Render strikethrough using one SGR sequence |
| [6a366f6](https://github.com/nospor/teams-tui-go/commit/6a366f6) | Configurable tenant ID for single-tenant app registrations |
| [2d05325](https://github.com/nospor/teams-tui-go/commit/2d05325) | Previews for forwarded-message attachments |
| [ff923ab](https://github.com/nospor/teams-tui-go/commit/ff923ab) | Avoid the short timeout for large attachment downloads |
| [a7c79c3](https://github.com/nospor/teams-tui-go/commit/a7c79c3) | Download-only shortcut and attachment status in the popup |
| [ae9a585](https://github.com/nospor/teams-tui-go/commit/ae9a585) | Prevent blank lines in message edit round-trips |

Some areas overlap with independent fork changes. Conflicts were resolved
around the fork's behavior rather than replacing complete files with upstream
copies. Automated tests cover these paths; live tenant sign-in and native
desktop launches still need validation on the user's machine.

To refresh the comparison from a checkout:

```bash
git fetch https://github.com/nospor/teams-tui-go.git main:refs/remotes/upstream/main
git log --oneline main..upstream/main
git rev-list --left-right --count main...upstream/main
```
