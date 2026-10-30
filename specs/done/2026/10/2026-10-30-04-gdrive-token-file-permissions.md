---
model: sonnet
effort: low
automerge: false
---

# The Google Drive OAuth token file is created world-readable

## Problem
Found by the security scan of 2026-10-30 (finding L2). Severity: Low. Any local user, process or backup job that can read the server's working directory can trigger it.

When the `gdrive` backend obtains a token, it saves it with `drvoa.StoreTokenToFile(tokenFile, auth.Token)` (`fs/gdrive/gdrive.go:84`). That helper (`github.com/fclairamb/afero-gdrive` `oauthhelper`, `StoreTokenToFile`) uses `os.Create`, so the file gets mode `0666` minus the umask, usually `0644`. By default the file is `gdrive_token_<user>.json` in the current directory (`fs/gdrive/gdrive.go:43-45`). It holds a refresh token with full Drive access to the account, valid until revoked.

## Proposal
1. In `fs/gdrive/gdrive.go`, write the token with a local helper: marshal it to JSON and `os.WriteFile(tokenFile, data, 0o600)`, then `os.Chmod(tokenFile, 0o600)` so that an existing file with a wider mode is also fixed (`WriteFile` keeps the mode of an existing file).
2. Optionally fix `StoreTokenToFile` in afero-gdrive the same way and bump it; the local helper is enough for this repo.

## Tests
- `fs/gdrive/gdrive_test.go` (new): the token-saving helper writes a token to a temp file whose mode is `0600`. Fails before the fix if written through `drvoa.StoreTokenToFile` (mode `0644` with the usual umask; set `syscall.Umask(0o022)` in the test for determinism, skip on Windows).
- Same file: a pre-existing token file with mode `0644` ends with mode `0600` after saving.
- Positive control: `drvoa.LoadTokenFromFile` reads the saved token back with the same access and refresh tokens.
