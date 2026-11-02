---
model: opus
effort: low
automerge: false
---

# The docker workflow still publishes images from the `chore/cleanup` branch

## Problem
Found by the security scan of 2026-11-02 (finding L1). Severity: Low. It can be triggered by anyone with write access to the repository.

`.github/workflows/docker.yml:19-24` still has the test trigger added for #1833, marked `TODO: remove before merging`: a push to the `chore/cleanup` branch runs the job with `PUSH=true` (`:43`) and publishes `chore-cleanup` and `chore-cleanup-alpine` images to Docker Hub and `ghcr.io/fclairamb/ftpserver` (`type=ref,event=branch`, `:82`). Images built from an unreviewed branch can then be published outside of the release process, under the project's official repositories.

## Proposal
1. Remove the `branches: [chore/cleanup]` entry and its TODO comment from the `push` trigger of `.github/workflows/docker.yml:19-24`, keeping only the `v*.*.*` tags.
2. Remove `type=ref,event=branch` from the metadata tags (`:82`), now unused.

## Tests
- `actionlint .github/workflows/docker.yml` passes.
- A pull request touching `Dockerfile` still runs the workflow (build only, no push), as a positive control.
