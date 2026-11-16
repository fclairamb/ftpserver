---
model: opus
effort: medium
automerge: false
---

# The first Google Drive authorization blocks the FTP login forever

## Problem
Found by the security scan of 2026-11-16 (finding L1). Severity: Low. Any FTP user who has the credentials of a
`gdrive` access whose token file doesn't exist yet (or can't be read) can trigger it. Introduced by b30ebd8 (#1885),
released in 0.16.7.

When no token can be loaded (`fs/gdrive/gdrive.go:53-61`), `LoadFs` calls `auth.NewHTTPClient(context.Background())`
(`:84`) without an `Authenticate` callback (`:70-82`). afero-gdrive v0.5.0 then runs the loopback flow: it opens a
listener on `127.0.0.1`, logs the authorization URL and waits for the redirect (`oauthhelper/helper.go:150-183`).
`waitForCode` (`:209-231`) only stops on a callback or when `ctx` is done, and `context.Background()` never is.

The FTP login runs synchronously in the client's control goroutine (ftpserverlib v0.33.2 `handle_auth.go:95` calls
`AuthUser`, which calls `s.loadFs`, `server/server.go:435`). So until someone completes the authorization in a
browser:
- each such login stays blocked forever, holding its goroutine, its control connection and a loopback port, even
  after the client goes away. `ClientDisconnected` never runs, so the client is never uncounted: with `max_clients`, a
  few logins lock every user out. `WaitGracefully` also waits for them at shutdown.
- for a `shared` access, `loadFs` calls the loader under the cache lock (`server/server.go:296-303`). Every login on
  any shared access, every disconnect from one (`releaseCachedFs`, `:326-328`) and every `SIGHUP` reload
  (`ReloadConfig`, `:156`) then block too.
- every login starts its own flow, with a new listener and a new URL in the logs.

Before b30ebd8, the code was read from stdin with `fmt.Scan`. Without a terminal (Docker image, systemd service),
that returned an error at once and only the login failed. Reproduced by the scan: `fs.LoadFs` on a `gdrive` access with
a missing token file is still blocked after 5 seconds.

## Proposal
1. `fs/gdrive/gdrive.go`: when no token was loaded, get it first with a bounded context:
   `context.WithTimeout(context.Background(), authorizationTimeout)` (5 minutes by default, overridable by an
   `authorization_timeout` param as a Go duration). Then build the HTTP client with a second `drvoa.Auth` holding that
   token and `context.Background()`. Don't pass the timeout context to the client: oauth2 uses that context for later
   token refreshes, which a cancelled context would break.
2. Allow one pending authorization per token file: a package-level map guarded by a mutex. A login that needs a token
   file whose authorization is already running fails right away with an explicit error ("Google Drive authorization
   pending"), without starting a second flow.
3. Return a clear login error when the timeout expires, and log it with the token file and user name.
4. `fs/gdrive/README.md`: say that the first authorization must be completed within the timeout, and that logins fail
   until then.

## Tests
- `fs/gdrive/gdrive_test.go`: `LoadFs` on an access with a missing `token_file` (temp dir), fake client id/secret and
  `authorization_timeout: "200ms"`: returns an error within 2 seconds. Fails before the fix (still blocked at the test
  deadline; run it in a goroutine with a `select` on a timer).
- Concurrency: while a first `LoadFs` waits (long timeout), a second one on the same token file returns the "pending"
  error at once. Positive control: a second one on another token file starts its own flow, and once the first one
  times out, a new login on the first token file can start a flow again.
- Positive control: with an existing, valid token file, `LoadFs` doesn't start the flow (no authorization URL logged).

## To verify
- Whether `drv.New(httpClient)` (afero-gdrive) calls the Drive API at creation. If it does, the positive control with a
  token file needs a fake HTTP transport, or it has to stop at checking that no flow started.
