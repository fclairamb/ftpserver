---
model: opus
effort: medium
automerge: false
---

# The shared file system cache hands one access's backend to other accesses

## Problem
Found by the security scan of 2026-11-02 (finding M1). Severity: Medium. Any user who can log in can trigger it, on a server where at least one access has `shared: true` and either several accesses share a user name, or the accesses come from a webhook that doesn't return a distinct `user` for each login.

`Server.loadFs` (`server/server.go:211-238`) looks up the cache by `access.User` alone (`server/server.go:222`), and does it **before** checking `access.Shared`. Only the store is limited to shared accesses (`server/server.go:232-235`). So once a shared access has been used, every later login whose access has the same `User` field gets that cached file system, whatever its own backend, parameters or `read_only` flag:

- **Config file**: accesses are evaluated in their declaration order (`config/config.go:251-283`), so the same user name can appear several times with different passwords, for example a read-only and a read-write access to two directories. After a login on the shared read-write access, a login with the read-only password gets the read-write file system of the other access.
- **Webhook** (`server/server.go:240-292`): `access.User` is the `user` field of the webhook's reply, not the name the client logged in with. A webhook that omits it, or returns the same value for everyone, makes every user get the file system of the first user whose reply had `shared: true`: read and write access to another tenant's directory or bucket. The README (`README.md:237-241`) only asks for "an access object" and doesn't say that `user` matters.

## Proposal
1. In `loadFs` (`server/server.go:211`), only use the cache when `access.Shared` is true: a non-shared access always gets a new file system.
2. Key the cache by the whole access, not by its user name: a fingerprint of the access (for example the SHA-256 of its JSON encoding, which includes `user`, `pass`, `fs`, `params`, `read_only`, `sync_and_delete` and `logging`). Two logins share an instance only when their accesses are identical. Keycloak accesses keep bypassing the cache (`server/server.go:214`).
3. For webhook accesses, also include the login name passed to `AuthUser` in the key (`server/server.go:295`), so a reply without `user` never shares an instance between two users. Simplest: `loadFs(user, access)`, with the key built from both.
4. README webhook section (`README.md:237-256`) and `fs/README.md:21-25`: say that a shared instance is reused for the same user and identical access only.

## Tests
- `server/server_internal_test.go`, config case: two accesses with the same user name and different passwords, the first `shared: true` on a writable temp dir, the second `read_only: true` on another temp dir. Log in with the first password, then the second: creating a file on the second session's file system must fail, and listing it shows the second dir's content. Fails before the fix.
- Webhook case: an `httptest` webhook returning `{"fs":"os","shared":true,"params":{"basePath":<dir of the user>}}` without `user`. Log in as alice then bob: bob's file system must be confined to bob's dir (a file created by alice isn't visible). Fails before the fix.
- Positive control: two logins of the same user on the same shared access get the same instance (cache hit), and a reload (`ReloadConfig`) still drops it.
