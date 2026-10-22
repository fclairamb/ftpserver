# Contributing

Issues and pull requests are welcome.

## Development

```sh
go build -v ./...        # Build
go test -race -v ./...   # Run the tests
golangci-lint run        # Lint (golangci-lint v2, configured by .golangci.yml)
go run . -conf ftpserver.json
```

If you change the config file format, update the [JSON schema](config-schema.json) and the documentation accordingly.
Sample config files (in `config/samples/` and `fs/*/`) are validated against the schema by the CI.

## Adding a backend

A backend is a package in `fs/` exposing a `LoadFs(access *confpar.Access, ...) (afero.Fs, error)` function, that
reads its parameters from `access.Params`. It must then be registered in [fs/fs.go](fs/fs.go), added to the `fs` enum
of the [JSON schema](config-schema.json), and documented in a `README.md` in its directory, linked from the
[main README](README.md#supported-backends).

## Pull requests

Pull request titles must follow the [conventional commits](https://www.conventionalcommits.org/) format
(`feat: ...`, `fix: ...`, `chore: ...`), as they are used to generate the [changelog](CHANGELOG.md) and the releases.
