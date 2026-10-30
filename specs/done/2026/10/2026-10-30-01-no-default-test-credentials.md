---
model: opus
effort: medium
automerge: false
---

# The auto-created configuration serves a writable `test`/`test` account on all interfaces

## Problem
Found by the security scan of 2026-10-30 (finding M1). Severity: Medium. Anyone who can reach port 2121 can trigger it, without any credential, on a server started without `-conf`.

When the server starts without `-conf` (`main.go:45-56`), it writes `ftpserver.json` from `confFileContent()` (`main.go:163`), which declares a `test` user with the password `test` on the `os` backend rooted at `/tmp`, read-write (`main.go:168-172`). The listen address defaults to `0.0.0.0:2121` (`config/config.go:188`) and plain FTP is allowed.

This is the default path of the published Docker images: their `ENTRYPOINT` passes no argument (`Dockerfile:31`, `Dockerfile.alpine:28`), and the README's run command mounts host directories on `/tmp` and `/app` (`README.md:84`). Anyone on the network who knows the documented credentials can read, write and delete the mounted host files, and plant files that other services read.

## Proposal
1. In `createConfFile` (`main.go:123`), generate a random password (crypto/rand, 16+ characters from an unambiguous alphabet) instead of `test`, and pass it to `confFileContent`.
2. Store it hashed with the existing bcrypt helper (`config/config.go`, `HashPlaintextPasswords` logic), or keep it plain in a `0o600` file (already the mode, `main.go:130`): see Open questions.
3. Log the user and the generated password once, at `Warn` level, when the file is created ("No conf file, creating one with user test and password ..."). Never log it on later starts.
4. Update `README.md:84-115` and `config/README.md` to read the password from the logs (`docker logs`) instead of `test:test`.
5. Update the Docker test workflow (`.github/workflows/docker-test.yml`), which logs in with `test:test`: provide its own config file with `-conf`, or read the password from the container logs.

## Tests
- `main_test.go` (new): `createConfFile` in a temp dir creates a file whose `test` access does not accept `test` as a password (`config.GetAccess("test", "test")` returns `ErrUnknownUser`). Fails before the fix.
- Same file, positive control: the password returned or logged by `createConfFile` is accepted by `GetAccess`, and two calls in two temp dirs produce different passwords.
- Same file: `createConfFile` does not touch an existing file.

## Open questions
- Keep a fixed, documented password for the quick start, and only print a warning? Recommended: no, generate a random one; a known password on a network service is the issue.
- Store the generated password hashed? Recommended: yes, bcrypt like `hash_plaintext_passwords`, so the config file never holds it in clear.

## Resolved open questions
- Generate a random password at first start; no fixed documented password (autopilot: recommended answer).
- Store the generated password hashed with bcrypt, like `hash_plaintext_passwords` (autopilot: recommended answer).
