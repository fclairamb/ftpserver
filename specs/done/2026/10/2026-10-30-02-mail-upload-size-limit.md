---
model: sonnet
effort: medium
automerge: false
---

# The mail backend keeps every upload in memory without any limit

## Problem
Found by the security scan of 2026-10-30 (finding M2). Severity: Medium. Any user who can log in on an access using the `mail` backend (including `anonymous` if configured) can trigger it.

`File.Write` appends every uploaded chunk to an in-memory slice (`fs/mail/mail.go:199-200`) and the whole content is only sent, as an attachment, when the file is closed (`fs/mail/mail.go:112-124`). Nothing bounds its size. One upload of a few gigabytes, or a few parallel ones, exhausts the server's memory: the process is killed and every user of every backend is disconnected. The SMTP server would refuse such a message anyway, but only after it has been fully buffered.

## Proposal
1. Add a `MaxSize` parameter to the mail backend (`fs/mail/mail.go:52`, `LoadFs`), in bytes, defaulting to 25 MiB (a common SMTP message limit). `0` or a negative value is rejected with `ErrInvalidParameter`.
2. Store it in `Fs` and make `File.Write` (`fs/mail/mail.go:199`) return an error (a new `ErrFileTooLarge`) once `len(f.Content)+len(b)` exceeds it, without appending. `Close` must then not send a truncated mail: keep a flag on the file and return the error.
3. Document `MaxSize` in `fs/mail/README.md` and in `config-schema.json` if the mail params are described there.

## Tests
- `fs/mail/mail_test.go`: with `MaxSize` 1024, writing 2048 bytes returns `ErrFileTooLarge` and the buffered content stays at most 1024 bytes. Fails before the fix (the write succeeds).
- Same file, positive control: writing 512 bytes in several chunks succeeds and the content is complete.
- Same file: `LoadFs` with `MaxSize` set to `abc`, `0` or `-1` returns `ErrInvalidParameter`; without `MaxSize` the default applies.
- `Close` after a refused write returns an error and does not call the dialer (inject a dialer or check before `DialAndSend`).

## To verify
- How the current tests build an `Fs` without a real SMTP server, to test `Close` without sending anything.
