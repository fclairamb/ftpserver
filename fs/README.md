# Common access options

On top of the backend (`fs`) and its `params`, every access supports a few options.

## Read-only

The `read_only` boolean blocks any write to the underlying file system.

```json
{
  "user": "test",
  "pass": "test",
  "fs": "os",
  "read_only": true,
  "params": {
    "basePath": "/tmp"
  }
}
```

## Shared

The `shared` boolean makes the backend instance created at the first login of the user reused by all their later
sessions, instead of creating a new one for each session. This is useful for backends with an expensive setup (SFTP
connection, Telegram bot...). Shared instances are dropped when the configuration is reloaded.

```json
{
  "user": "test",
  "pass": "test",
  "fs": "os",
  "shared": true,
  "params": {
    "basePath": "/tmp"
  }
}
```

## Sync & Delete

The `sync_and_delete` option makes uploaded files written to a local temporary directory first, and then moved to the
backend once they are complete (see [afero-snd](https://github.com/fclairamb/afero-snd)). The `directory` is optional,
a system temporary directory is used when it isn't specified.

```json
{
  "user": "test",
  "pass": "test",
  "fs": "os",
  "sync_and_delete": {
    "enable": true,
    "directory": "/tmp/snd"
  },
  "params": {
    "basePath": "/target"
  }
}
```

## Logging

The `logging` option enables the logging of the FTP exchanges and/or the file accesses of this access only. The same
options exist at the server level to enable them for all the accesses.

```json
{
  "user": "test",
  "pass": "test",
  "fs": "os",
  "logging": {
    "ftp_exchanges": true,
    "file_accesses": true
  },
  "params": {
    "basePath": "/target"
  }
}
```
