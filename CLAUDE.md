# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- `go build -v ./...` - Build all packages
- `go test -race -v ./...` - Run all tests with race detection
- `golangci-lint run` - Lint (golangci-lint v2, configured by `.golangci.yml`, run by the CI)
- `go run . -conf ftpserver.json` - Run with a specific config file (`ftpserver.json` is created if no `-conf` is given)
- `go run . -conf-only` - Only create/check the configuration file

## Architecture

A Go FTP server acting as a gateway between FTP clients and various storage backends. It's built on top of
[ftpserverlib](https://github.com/fclairamb/ftpserverlib) (the FTP protocol) and uses the
[afero](https://github.com/spf13/afero) `Fs` interface for the backends.

- `main.go` - Entry point: config loading, server instantiation, signals (`SIGHUP` reloads the config, `SIGTERM`
  stops gracefully).
- `config/` - Config file loading and user authentication (plain-text and hashed passwords).
  `config/confpar/` holds the config structures, which must stay in sync with `config-schema.json`.
- `server/` - The ftpserverlib driver: settings, client counting, authentication (config or webhook), shared file
  system cache, TLS.
- `fs/` - `fs.go` instantiates the backend of an access and applies the common options (read-only, sync & delete).
  Each subdirectory is a backend exposing a `LoadFs` function: `afos` (local disk), `s3`, `gcs`, `dropbox`, `gdrive`,
  `sftp`, `mail`, `telegram`, `keycloak`. `fslog` wraps a file system to log accesses, `utils` has shared helpers.

## Conventions

- Backend parameters are strings read from `access.Params`. Document them in the backend's `README.md` and keep the
  JSON schema and the main README backend table up to date.
- Imports: standard library, third-party, then `github.com/fclairamb/ftpserver/...` packages.
- Line length limit: 120 characters.
- Pull request titles follow conventional commits; releases are cut by release-please, which also triggers the
  binaries (`release-artifacts.yml`) and docker images (`docker.yml`) builds.
- `build.go` holds the version variables, set at build time with `-ldflags -X`.

## Automerge

A PR labelled `automerge` merges on its own once CI is green: `.github/workflows/auto-merge.yml` enables GitHub's
auto-merge (squash) and updates a branch that is behind main, with the `CI_PAT` secret (a branch update pushed with
`GITHUB_TOKEN` would not trigger CI). Renovate's minor/patch updates get the label.
