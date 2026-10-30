---
model: sonnet
effort: medium
automerge: false
---

# The Docker images run the server as root

## Problem
Found by the security scan of 2026-10-30 (finding L1). Severity: Low (hardening). It matters once someone can exploit a bug of the server or of a backend library: anyone who can connect to the FTP port, with the default credentials of finding M1.

Neither image sets a `USER`: `Dockerfile:24-31` (scratch) and `Dockerfile.alpine:24-28` run `/bin/ftpserver` as uid 0. Combined with the bind mounts the README recommends (`README.md:84`: `-v $(pwd)/files:/tmp -v $(pwd):/app`), files written over FTP land on the host owned by root, and a code execution bug in the process runs as root in the container.

## Proposal
1. In both Dockerfiles, run as an unprivileged user, for example `USER 65532:65532` (the distroless "nonroot" ids; scratch has no `/etc/passwd`, numeric ids work).
2. Make `/app` (where `ftpserver.json` is auto-created, `main.go:123-134`) and `/tmp` (served by the default config) writable by that user: create them in the builder stage with `install -d -o 65532 -g 65532` and copy them with `COPY --chown=65532:65532`, keeping the sticky bit on `/tmp`.
3. Ports 2121-2130 are above 1024, no capability is needed. Check that `.github/workflows/docker-test.yml` still passes (upload and download through the image).
4. Document in `README.md` that the mounted directories must be writable by uid 65532, or that `--user` can override it.

## Tests
- `.github/workflows/docker-test.yml`: add a step checking that `docker inspect --format '{{.Config.User}}' <image>` is `65532:65532` (scratch has no shell to run `id`). Fails before the fix (empty user).
- Same workflow, positive control: the existing upload/download test still passes with the new user, for the scratch and alpine images.

## Open questions
- Is a breaking change acceptable for users who mount host directories owned by another uid? Recommended: yes, with a note in the changelog and the `--user` override documented.

## Resolved open questions
- A breaking change for host mounts owned by another uid is acceptable, with a changelog note and the `--user` override documented (autopilot: recommended answer).
