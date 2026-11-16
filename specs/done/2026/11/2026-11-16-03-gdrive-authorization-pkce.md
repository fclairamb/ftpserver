---
model: opus
effort: medium
automerge: false
---

# Bind the Google Drive authorization code to the server with PKCE

## Problem
Found by the security scan of 2026-11-16 (finding L2). Severity: Low. Someone who can both read the server's logs and
reach its loopback interface during the first authorization of a `gdrive` access can trigger it: a local account or a
co-hosted process, for example. Introduced by b30ebd8 (#1885), released in 0.16.7.

`fs/gdrive/gdrive.go:72-81` logs the whole authorization URL at Warn level. It includes the `state` and the
`redirect_uri` (`http://127.0.0.1:<port>/`) of the pending flow. afero-gdrive v0.5.0's loopback server accepts the
first request that carries this `state` and a `code` (`oauthhelper/helper.go:234-255`), then exchanges the code
(`:178`). The flow uses no PKCE (`AuthCodeURL` at `:165` has no `code_challenge`; the scan saw none in the logged URL).
So nothing ties the code to the authorization request the server started.

Someone who has the logged URL can run their own consent flow with the same public client id and loopback redirect,
then hand the resulting code to the server's listener with the logged `state`. The server stores a refresh token for
**their** Drive account in the token file. From then on, the files that FTP users upload to that access go to the
attacker's Drive, and the files they download come from it. The old flow read the code from the server's stdin, which
needed a console on the server.

## Proposal
1. afero-gdrive (`github.com/fclairamb/afero-gdrive`, same author), `oauthhelper/helper.go`: generate a verifier with
   `oauth2.GenerateVerifier()` for each flow, add `oauth2.S256ChallengeOption(verifier)` to `AuthCodeURL` (`:165`, and
   `:131` for the `Authenticate` callback path) and `oauth2.VerifierOption(verifier)` to `Exchange` (`:178`, `:138`).
   A code obtained from another authorization request then fails the exchange.
2. To make it testable, let the tests set the OAuth endpoint (today hardcoded in `NewHTTPClient`, `:84-87`): an unexported
   field or an option.
3. Release afero-gdrive and bump it in `go.mod`.
4. Optional, in `fs/gdrive/gdrive.go:72-81`: say in the log message that the URL must not be shared.

## Tests
- afero-gdrive `oauthhelper`: with a fake authorization/token server (`httptest`), the authorization URL contains
  `code_challenge` and `code_challenge_method=S256`, and the token request carries a `code_verifier` whose S256 hash
  is that challenge. Fails before the fix (no challenge in the URL, no verifier in the request).
- The fake token server only accepts a code with the verifier it was issued for: a code issued for another challenge
  is refused, and the flow returns an error without storing a token.
- Positive control: the normal flow (callback with the right `state` and its own code) completes and returns the token.
- ftpserver: after the bump, `go test ./fs/gdrive/...` still passes.
