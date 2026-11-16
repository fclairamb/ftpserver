---
model: opus
effort: medium
automerge: false
---

# SITE CHMOD on an s3 access changes the object's ACL and can make it public

## Problem
Found by the security scan of 2026-11-16 (finding M1). Severity: Medium. Any authenticated FTP user with write access
(not `read_only`) on an `s3` access can trigger it. Files can also be published by accident: some FTP clients send
`SITE CHMOD` after each upload to keep the local permissions.

`s3.LoadFs` hands the afero-s3 file system to the user as is (`fs/s3/s3.go:66`, wrapped only by `BasePathFs` at
`:73`). ftpserverlib's `SITE CHMOD` calls the driver's `Chmod` with the user's octal mode
(ftpserverlib v0.33.2 `handle_files.go:277-301`). afero-s3 (since v0.3.0, v0.5.1 `s3_fs.go:317-339`) maps that mode to
a canned ACL and calls `PutObjectAcl`:
- "other" read and write bits set (`666`, `777`...): `public-read-write`
- "other" read bit set (`644`, `755`...): `public-read`
- otherwise: `private`

So with a mode as common as `644`, the object becomes readable by anyone on the internet who knows its URL. Reproduced
by the scan with a local fake S3 endpoint: `Chmod("/report.pdf", 0644)` through `fs.LoadFs` sends
`PUT /acme/users/u/report.pdf?acl` with `x-amz-acl: public-read`.

Impact:
- A user can turn the operator's bucket into a public file host for any content they upload, under the operator's
  domain and bandwidth.
- A user can publish any file of the access's prefix, including files uploaded by other users of the same access.
- Clients that keep permissions after an upload publish every file they send, without the user noticing.
- `private` replaces the object's ACL, which drops any grant the operator set on it.

It only takes effect when the bucket accepts ACLs, its public access isn't blocked, and the server's credentials
allow `s3:PutObjectAcl`. Buckets created with the current defaults (ACLs disabled, public access blocked) refuse the
call, but older buckets and several S3-compatible stores accept it. Affected: ftpserver 0.9.0 (afero-s3 v0.3.0) to
0.16.7. The other backends aren't affected: the GCS one refuses `Chmod`, and on `os` and `sftp` it is a plain file mode
inside the base path.

## Proposal
1. `fs/s3/s3.go`: wrap the afero-s3 file system (before the `BasePathFs` at `:73`) in a small type embedding
   `afero.Fs` whose `Chmod` returns an error (for example `ErrChmodNotSupported`, "chmod is not supported on s3")
   without calling the backend. Everything else is passed through.
2. Add an opt-in access param `chmod_acl` (`"true"`): when set, keep the current behaviour (mode → canned ACL), for
   operators who rely on `SITE CHMOD` to publish files on purpose.
3. Document `chmod_acl` in `fs/s3/README.md` (with a warning that it lets users make objects public), in
   `config-schema.json` (s3 params) and in the main README backend table if it lists the s3 params.
4. Changelog / release note: operators of an s3 access whose bucket accepts ACLs should check that no object was made
   public through `SITE CHMOD`.

## Tests
- `fs/s3/s3_test.go`: start an `httptest` server acting as the S3 endpoint (path style, static credentials) that
  records every request with an `acl` query parameter. `LoadFs` without `chmod_acl`, then `Chmod("/f", 0o644)`:
  returns an error and the server received no ACL request. Fails before the fix (a `PUT ...?acl` with
  `x-amz-acl: public-read` is sent and `Chmod` returns nil).
- Positive control: the same access with `chmod_acl: "true"`: `Chmod(0o644)` sends `public-read` and `Chmod(0o600)`
  sends `private`.
- Other operations still reach the backend through the wrapper: for example `Stat` or `Remove` sends its request to the
  fake endpoint.

## Open questions
- Refuse `SITE CHMOD` with an error (550), or accept it silently without effect? Recommended: refuse it. A silent
  success misleads the user, and FTP clients treat a failed `SITE CHMOD` after an upload as a warning, not a failed
  upload.
