---
model: sonnet
effort: low
automerge: false
---

# The release workflow runs a third-party action pinned by a mutable tag with write access

## Problem
Found by the security scan of 2026-10-30 (finding L3). Severity: Low. Whoever controls the third-party action's repository (its maintainer, or an attacker who takes it over) can trigger it, at the next release.

`.github/workflows/release-artifacts.yml:67` uses `wangyoucao577/go-release-action@v1.55`, a tag that its owner can move, in a job holding `contents: write` (`release-artifacts.yml:26-27`) and the `GITHUB_TOKEN` (`:69`), with `overwrite: true` on release assets. A moved tag would run arbitrary code that builds and uploads the release binaries, and could replace the assets of existing releases.

## Proposal
1. Pin `wangyoucao577/go-release-action` to the full commit SHA of `v1.55`, with the tag in a trailing comment (`uses: wangyoucao577/go-release-action@<sha> # v1.55`).
2. Do the same for the other non-GitHub third-party actions (`amannn/action-semantic-pull-request`, `.github/workflows/semantic-pr.yml:18`; `docker/*` and `golangci/*` are optional).
3. Enable digest pinning updates in `.github/renovate.json` (`helpers:pinGitHubActionDigests`) so that Renovate keeps the SHAs current.

## Tests
- No unit test: check with `grep -nE 'uses: [^.][^@]*@v[0-9]' .github/workflows/*.yml` that no third-party action outside `actions/`, `github/`, `docker/` and `golangci/` is pinned by tag. Matches before the fix, none after.
- Positive control: the release workflow still parses (`actionlint` if available) and Renovate's config validates (`npx --yes renovate-config-validator`).
