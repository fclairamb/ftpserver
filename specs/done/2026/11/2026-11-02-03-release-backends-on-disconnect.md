---
model: opus
effort: high
automerge: false
---

# SFTP and Telegram backends leak a connection on every login

## Problem
Found by the security scan of 2026-11-02 (finding M3). Severity: Medium. Any user who can log in on an `sftp` or `telegram` access without `shared: true` can trigger it.

ftpserverlib never closes a client driver when the session ends, and `ClientDisconnected` (`server/server.go:160-173`) doesn't release the file system created by `AuthUser` (`server/server.go:312`). For the backends that hold a connection, each login leaks it for the life of the process:

- `sftp`: `LoadFs` dials an SSH connection and opens an SFTP client (`fs/sftp/sftp.go:89-99`) that are never closed: one TCP connection to the SFTP host and its goroutines per FTP login.
- `telegram`: `LoadFs` creates a bot and starts its long poller in a goroutine (`fs/telegram/telegram.go:158-177`) that is never stopped: one goroutine polling the Telegram API every 10 seconds per FTP login. Several pollers on the same bot token also make Telegram reject the polls.

Without `shared`, a new backend is created at each login (`server/server.go:228-237`). A user who logs in and out in a loop piles up connections until the process runs out of file descriptors or memory, or the SFTP server refuses new sessions (its `MaxStartups`/`MaxSessions` limits), so every user of the server is cut off. Normal use leaks too, more slowly. Shared instances leak the same way when a reload (`server/server.go:120-131`) drops the cache.

## Proposal
1. `fs/sftp/sftp.go`: return a file system that also implements `io.Closer`, closing the SFTP client and the SSH connection (a small wrapper embedding `*sftpfs.Fs`).
2. `fs/telegram/telegram.go`: add `func (m *Fs) Close() error` that stops the bot (`bot.Stop()`), and make the poller start and stop safe to call once.
3. The wrappers applied in `fs/fs.go:60-77` (`ReadOnlyFs`, sync & delete) and `fslog` (`server/server.go:332`) hide the backend's `Close`: keep a reference to the backend's closer next to the wrapped file system (for example `fs.LoadFs` returning it, or a small `closableFs` type).
4. `server/server.go`: in `AuthUser`, for a non-shared access, remember the closer by `cc.ID()`; in `ClientDisconnected`, close it and forget it. In `ReloadConfig`, close the dropped shared instances once no session uses them anymore (or document that they are left to the existing sessions and closed at their end: reference counting per cached instance).
5. Keycloak and the other backends without a connection need nothing.

## Tests
- `server/server_internal_test.go`: register a fake backend (or use a test hook in `loadFs`) whose file system counts `Close` calls. Authenticate a non-shared access, then call `ClientDisconnected` with the same client context: `Close` is called once. Fails before the fix.
- Shared access: two sessions, disconnect both, `Close` is not called; after `ReloadConfig` and the end of the sessions, it is called once.
- `fs/sftp/sftp_test.go`: with an in-process SSH/SFTP server (`golang.org/x/crypto/ssh` + `github.com/pkg/sftp` `NewServer`), after `Close` the server sees the connection closed.
- `fs/telegram`: `Close` stops the poller (fake poller or `httptest` Telegram API), and a second `Close` doesn't panic.

## Open questions
- Close dropped shared instances at reload (reference counting), or leave them as they are? Recommended: reference counting, otherwise each reload of a server with a shared SFTP access leaks one connection.

## Resolved open questions
- Close dropped shared instances at reload using reference counting (autopilot: recommended answer), otherwise each reload leaks one connection.
