---
model: opus
effort: high
automerge: false
---

# The sync & delete wrapper is never closed, and its pending copies run after the backend is closed

## Problem
Found by the security scan of 2026-11-09 (finding L1). Severity: Low. Any user who can log in on an access with
`sync_and_delete.enable: true` can trigger the leak; the lost uploads happen in normal use. It is a gap in the fix of
the 2026-11-02 finding M3 (`specs/done/2026/11/2026-11-02-03-release-backends-on-disconnect.md`, #1857).

`fs.LoadFs` takes the closer from the backend before wrapping it (`fs/fs.go:69`), then wraps it with `snd.NewFs`
(`fs/fs.go:83`). The closer it returns (`fs/fs.go:98`) is the backend's only, so nothing ever calls the wrapper's own
`Close`. Two consequences:

1. **Leak per login.** afero-snd v0.2.0 `NewFs` (`fs.go:61`) starts a goroutine (`fs.go:160`) with a ticker and a
   1000-slot operation queue. When `directory` is empty it also creates a directory with `os.MkdirTemp` (`fs.go:68`).
   Only `Close` (`fs.go:164-166`) stops them, and it never removes that directory. A non-shared access creates one
   instance per login (`server/server.go:278-287`), so a user who logs in and out in a loop piles up goroutines,
   tickers and, without `directory`, empty `/tmp/afero-snd*` directories, until the process is restarted. Each leaked
   instance also keeps its backend referenced, and its ticker keeps running the garbage collector on its temporary
   directory every 10 minutes. With the `os` backend, the closer is nil, so nothing is released at all. Shared
   instances leak one wrapper per reload (`server/server.go:161-169`). Reproduced by the scan: 50 calls to
   `fs.LoadFs` on an `os` access with `sync_and_delete`, each followed by closing the returned closer, leave 50 extra
   goroutines.
2. **Uploads lost on SFTP and Telegram.** The wrapper copies a file to the backend asynchronously: its `File.Close`
   only queues the copy (afero-snd `file.go:28`), and the client already gets its `226`. Since #1857,
   `ClientDisconnected` releases the session (`server/server.go:204`), which closes the SFTP connection or the
   Telegram poller right away (`server/server.go:287`, and `:326` for a dropped shared instance). A copy still in the
   queue then runs against a closed backend and fails. The error isn't even logged: `fs/fs.go:86` passes no logger,
   so afero-snd discards it. The file stays in the temporary directory, where the garbage collector deletes it once
   it is older than 20 minutes and not among the 10 newest files (afero-snd `gc.go:27-80`). So a client that uploads
   then quits can lose its last files without any error.

## Proposal
1. afero-snd (same author, `github.com/fclairamb/afero-snd`): make `Close` wait until the operations queued before it
   have run (the goroutine exits after processing them, `Close` waits for it), with a bounded wait or a context so a
   hung backend can't block a disconnect forever. Make it safe to call twice. Remove the directory created by
   `os.MkdirTemp` once the queue is drained, never a `directory` set by the operator. Release a new afero-snd version
   and bump it in `go.mod`.
2. `fs/fs.go:60-98`: when the wrapper is used, return a closer that closes the wrapper first (so its pending copies
   reach the backend), then the backend's closer if there is one. Return it even when the backend has no closer
   (`os`, `s3`, ...). Keep it nil only when neither the wrapper nor the backend holds anything.
3. `server/server.go`: nothing to change, `loadFs` already calls the closer at the end of a session and for dropped
   shared instances. Check that the closer still runs outside the cache lock (`releaseCachedFs`), since it may now
   wait for copies.
4. `fs/README.md` "Sync & Delete" section: say that pending copies finish when the session ends.

## Tests
- `fs/fs_test.go` (new): `LoadFs` on an `os` access with `sync_and_delete` (temp `directory`), 50 times, closing the
  returned closer each time: `runtime.NumGoroutine()` returns to its starting value (allow a short wait). Fails before
  the fix: the closer is nil and 50 goroutines remain. Positive control: the same loop without `sync_and_delete`
  doesn't change the goroutine count either, and the returned closer is nil.
- Data test: wrap a backend that implements `io.Closer` and records the order of calls (a small test helper, for
  example by extracting the wrapping part of `LoadFs` into a function taking the backend). Write a file through the
  returned file system, close the file, then close the returned closer right away: the file is in the backend and the
  backend was closed after the copy. Fails before the fix (the backend is closed before the queued copy runs).
- afero-snd: `Close` runs the queued operations before returning, a second `Close` doesn't block or panic, and the
  `MkdirTemp` directory is removed while an operator-set `directory` is kept.

## Open questions
- Fix the waiting in afero-snd, or only in ftpserver (a wrapper closer that drains the queue itself is not possible
  without afero-snd exposing it)? Recommended: in afero-snd, then bump it here.

## Resolved open questions
- Fix the waiting in afero-snd, then bump it in ftpserver. (autopilot: recommended answer)
