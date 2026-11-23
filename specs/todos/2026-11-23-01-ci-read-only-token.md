---
model: opus
effort: low
automerge: false
---

# CI workflows that run dependency code get the repository's default write token

## Problem
Found by the security scan of 2026-11-23 (finding L1). Severity: Low. It can be triggered by whoever controls a new
release of one of the Go modules or PyPI packages the CI installs (an upstream compromise); no access to the
repository or to an FTP server is needed.

The repository's default workflow token is read-write (`default_workflow_permissions: write`, with
`can_approve_pull_request_reviews: true`), and four workflows don't declare `permissions:`, so they run with that
default token:

- `.github/workflows/build.yml:3-5` (every push on every branch, and pull requests): `go test -race -v ./...`
  (`:28`) compiles and runs the code of every dependency, `init` functions included.
- `.github/workflows/jsonschema.yml:3-5` (same triggers): `pip install jsonschema==4.26.0` (`:15`) also installs its
  transitive dependencies (`attrs`, `referencing`, `rpds-py`, `jsonschema-specifications`), unpinned.
- `.github/workflows/codeql-analysis.yml` and `.github/workflows/docker-test.yml`: no third-party code of
  ours runs there, but they also get the write token.

`actions/checkout` keeps the token in `.git/config` (default `persist-credentials: true`), so any code a step runs can
read it. Renovate pushes its branches to this repository, so a dependency update triggers `build.yml` on `push` with
the write token before anyone looks at the PR (an automerge label or a review changes nothing here).

Impact if a dependency release is malicious: with `packages: write` it can push an image to
`ghcr.io/fclairamb/ftpserver` (the package grants this repository the Write role, see `CLAUDE.md`), which third
parties pull; with `contents: write` it can replace the binaries attached to existing GitHub releases or push
branches and tags; with `pull-requests: write` it can approve PRs. Docker Hub isn't reachable (its secret is only
given to `docker.yml`). The workflows that do need write access (`docker.yml:28-30`, `release-please.yml`,
`release-artifacts.yml:26`, `auto-merge.yml:37`) already declare their permissions.

## Proposal
1. Add a top-level `permissions: contents: read` to `.github/workflows/build.yml`, `.github/workflows/jsonschema.yml`
   and `.github/workflows/docker-test.yml`.
2. In `.github/workflows/codeql-analysis.yml`, declare what CodeQL needs: `actions: read`, `contents: read`,
   `security-events: write`.
3. Set `persist-credentials: false` on the `actions/checkout` steps of those four workflows: none of them pushes.
4. Add a Go test (e.g. `workflows_test.go` in the root package, using `go.yaml.in/yaml/v3`, already in `go.mod`)
   that parses every `.github/workflows/*.yml` and fails when a workflow has no top-level `permissions:` and one of
   its jobs has no job-level `permissions:` either. Keeps a new workflow from silently getting the default token.
5. Mention in the PR description that the maintainer should also switch the repository setting *Actions → General →
   Workflow permissions* to "Read repository contents" and untick "Allow GitHub Actions to create and approve pull
   requests" (not settable from a file). Check first that `auto-merge.yml` and `release-please.yml` don't rely on the
   default: they use `CI_PAT` or declare their permissions.

## Tests
- The new workflow test fails on the current `main` (it names `build.yml`, `jsonschema.yml`, `codeql-analysis.yml`
  and `docker-test.yml`) and passes after the change.
- Positive control: the same test passes for `docker.yml`, `release-please.yml`, `release-artifacts.yml`,
  `auto-merge.yml` and `semantic-pr.yml` unchanged, and a fixture workflow with only job-level permissions passes.
- `actionlint` passes on the modified workflows; on the PR, `build`, `json-schema-check` and CodeQL still succeed
  (CodeQL uploads its results with the scoped token).

## To verify
- Whether `go.yaml.in/yaml/v3` is importable from the root package without promoting it to a direct requirement in
  a way `go mod tidy` would object to (otherwise parse with a minimal line-based check).
