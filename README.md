# Golang FTP Server

[![Build](https://github.com/fclairamb/ftpserver/actions/workflows/build.yml/badge.svg)](https://github.com/fclairamb/ftpserver/actions/workflows/build.yml)
[![Docker images](https://github.com/fclairamb/ftpserver/actions/workflows/docker.yml/badge.svg)](https://github.com/fclairamb/ftpserver/actions/workflows/docker.yml)
[![Docker test](https://github.com/fclairamb/ftpserver/actions/workflows/docker-test.yml/badge.svg)](https://github.com/fclairamb/ftpserver/actions/workflows/docker-test.yml)
[![Docker pulls](https://img.shields.io/docker/pulls/fclairamb/ftpserver)](https://hub.docker.com/r/fclairamb/ftpserver)
[![Go Reference](https://pkg.go.dev/badge/github.com/fclairamb/ftpserver.svg)](https://pkg.go.dev/github.com/fclairamb/ftpserver)

A gateway between old-school FTP devices (cameras, scanners, printers, legacy software...) and modern cloud based
file systems.

It is built on top of [ftpserverlib](https://github.com/fclairamb/ftpserverlib), for which it is the reference
implementation, and of the [afero](https://github.com/spf13/afero) `Fs` interface.

## Supported backends

| Backend | `fs` value | Documentation |
| --- | --- | --- |
| Local disk | `os` | [below](#local-disk) |
| [Amazon S3](https://aws.amazon.com/s3/) and S3-compatible storages (MinIO, Ceph...), through [afero-s3](https://github.com/fclairamb/afero-s3) | `s3` | [fs/s3](fs/s3/README.md) |
| [Google Cloud Storage](https://cloud.google.com/storage), through afero's [gcsfs](https://github.com/spf13/afero/tree/master/gcsfs) | `gcs` | [fs/gcs](fs/gcs/README.md) |
| [Dropbox](https://www.dropbox.com/), through [afero-dropbox](https://github.com/fclairamb/afero-dropbox) | `dropbox` | [fs/dropbox](fs/dropbox/README.md) |
| [Google Drive](https://developers.google.com/drive), through [afero-gdrive](https://github.com/fclairamb/afero-gdrive) | `gdrive` | [fs/gdrive](fs/gdrive/README.md) |
| [SFTP](https://en.wikipedia.org/wiki/SSH_File_Transfer_Protocol), through afero's [sftpfs](https://github.com/spf13/afero/tree/master/sftpfs) | `sftp` | [fs/sftp](fs/sftp/README.md) |
| Email (each uploaded file is sent as an attachment), thanks to [@x-way](https://github.com/x-way) | `mail` | [fs/mail](fs/mail/README.md) |
| Telegram (each uploaded file is sent to a chat), thanks to [@slayer](https://github.com/slayer) | `telegram` | [fs/telegram](fs/telegram/README.md) |
| Local disk with users authenticated by [Keycloak](https://www.keycloak.org/) | `keycloak` | [fs/keycloak](fs/keycloak/README.md) |

On top of any backend, an access can be made read-only, share its backend instance between sessions, write through a
local temporary directory or log every file access (see [fs/README.md](fs/README.md)).

## Features

FTP protocol support is brought by [ftpserverlib](https://github.com/fclairamb/ftpserverlib):

* Uploading and downloading files, with resume support (`REST`)
* Directory listing (`LIST`, `NLST`, [`MLST`/`MLSD`](https://tools.ietf.org/html/rfc3659#page-23))
* File and directory deletion and renaming, symbolic links (`SITE SYMLINK`, on the local disk backend)
* Passive (`PASV`, `EPSV`) and active (`PORT`, `EPRT`) transfers, IPv6 support ([RFC 2428](https://tools.ietf.org/html/rfc2428))
* Explicit TLS ([`AUTH TLS`](https://tools.ietf.org/html/rfc4217#section-4.1), [`PROT`](https://tools.ietf.org/html/rfc2228#page-8)) and implicit TLS
* File modification time ([`MDTM`](https://tools.ietf.org/html/rfc3659#page-8)) and size ([`SIZE`](https://tools.ietf.org/html/rfc3659#page-11))
* File hashing (`HASH`, `XCRC`, `MD5`, `XSHA256`...), when enabled
* Small memory footprint

On top of it, this server provides:

* Multiple users, each one with their own backend
* Hashed passwords (bcrypt, md5crypt, sha256crypt, sha512crypt), plain-text passwords can be hashed automatically
* Authentication through an HTTP webhook
* Configuration hot-reload (`SIGHUP`) and graceful shutdown (`SIGTERM`)

## Getting started

### Download it

Fetch a binary for your platform from the [latest release](https://github.com/fclairamb/ftpserver/releases/latest)
and run it. Archives are named `ftpserver-<version>-<os>-<arch>` and are provided for Linux, macOS and Windows on
`amd64` and `arm64`, plus `armv6`/`armv7` for Linux (Raspberry Pi and friends).

### Build it

```sh
go install github.com/fclairamb/ftpserver@latest
ftpserver
```

### Run it with docker

Images are available on [Docker Hub](https://hub.docker.com/r/fclairamb/ftpserver) and on the
[GitHub container registry](https://github.com/fclairamb/ftpserver/pkgs/container/ftpserver)
(`ghcr.io/fclairamb/ftpserver`) for `linux/amd64`, `linux/arm64`, `linux/arm/v6` and `linux/arm/v7`:

* `latest`, `<version>`: minimal image, containing only the binary
* `alpine`, `<version>-alpine`: image based on alpine, if you need a shell

The image runs from the `/app` directory, which is where it reads (or creates) its `ftpserver.json` config file. The
default config file serves the `/tmp` directory, with a `test` user whose random password is printed in the logs
(`docker logs <container>`) when the file is created.

```sh
# Creating a directory
mkdir -p files

# Starting the sample FTP server
docker run --rm -d -p 2121-2130:2121-2130 -v $(pwd)/files:/tmp -v $(pwd):/app fclairamb/ftpserver
```

### Run it with docker compose

```yml
# docker-compose.yml
services:
  ftpserver:
    image: fclairamb/ftpserver
    ports:
      - '2121-2130:2121-2130'
    volumes:
      - ./files:/tmp
      - .:/app
```

```sh
docker compose up -d
```

### Test it

The password of the `test` user is generated when the config file is created, and only printed then:

```sh
# Find the generated password in the logs
docker logs <container> 2>&1 | grep "No conf file"
PASS=<the password from the logs>

# Create some file
head -c 1048576 /dev/urandom > file.bin

# Upload it to the server
curl -v -T file.bin "ftp://test:${PASS}@localhost:2121/"

# Download it back
curl "ftp://test:${PASS}@localhost:2121/file.bin" -o file2.bin

# Compare it
cmp file.bin file2.bin
```

## Usage

```
ftpserver [-conf <file>] [-conf-only]
```

* `-conf <file>`: config file to use. When not specified, `ftpserver.json` is used, and created if it doesn't exist.
* `-conf-only`: only create the config file (if it doesn't exist) and check it can be loaded, then exit.

Signals:

* `SIGHUP` reloads the config file.
* `SIGTERM` stops accepting new connections and waits up to one minute for the connected clients to leave.

## Configuration

The configuration is a JSON file, described by a [JSON schema](config-schema.json). Referencing it with the
`$schema` property gives you auto-completion and validation in most editors. If you start the server without any
config file, a sample one is created with a single `test` user serving `/tmp`. Its password is random: it is
printed once in the logs when the file is created, and stored hashed (bcrypt) in the file.

```json
{
  "$schema": "https://raw.githubusercontent.com/fclairamb/ftpserver/main/config-schema.json",
  "version": 1,
  "listen_address": "0.0.0.0:2121",
  "passive_transfer_port_range": {
    "start": 2122,
    "end": 2130
  },
  "accesses": [
    {
      "user": "test",
      "pass": "test",
      "fs": "os",
      "params": {
        "basePath": "/tmp"
      }
    },
    {
      "user": "s3",
      "pass": "$2a$10$jG7tuqIlcUDMl1m1Ytj1TunU7pk.ko8lj3nOGzZvkIeU/BsfPVBra",
      "fs": "s3",
      "read_only": true,
      "params": {
        "bucket": "my-bucket",
        "region": "eu-west-1"
      }
    }
  ]
}
```

More samples are available in [config/samples](config/samples).

### General options

| Property | Description | Default |
| --- | --- | --- |
| `listen_address` | Address to listen on | `0.0.0.0:2121` |
| `public_host` | Public IP address to advertise for passive transfers, when the server is behind a NAT gateway. Can be overridden with the `PUBLIC_HOST` environment variable | |
| `passive_transfer_port_range` | Ports to use for passive transfers (`start`, `end`) | any port |
| `max_clients` | Maximum number of simultaneous clients, `0` means no limit | `0` |
| `idle_timeout` | Duration after which idle connections are closed (e.g. `5m`), a negative value disables it | `15m` |
| `tls` | TLS certificate files: `{"server_cert": {"cert": "cert.pem", "key": "key.pem"}}` | |
| `tls_required` | `ClearOrEncrypted`, `MandatoryEncryption` (explicit TLS required) or `ImplicitEncryption` (implicit FTPS) | `ClearOrEncrypted` |
| `hash_plaintext_passwords` | Replace the plain-text passwords of the config file by their bcrypt hash | `false` |
| `extensions.enable_hash` | Enable the `HASH` command and its variants | `false` |
| `logging.ftp_exchanges` | Log all the FTP commands and responses | `false` |
| `logging.file_accesses` | Log all the file accesses | `false` |
| `logging.file` | Also write the logs to this file | |
| `accesses` | The users and their backend, see below | |
| `accesses_webhook` | Authenticate users with a webhook instead of the `accesses`, see below | |

### Accesses

Each access defines a user and the backend serving their files:

| Property | Description |
| --- | --- |
| `user` | User name |
| `pass` | Password: plain-text or hashed (`$2a$`/`$2b$`/`$2y$` bcrypt, `$1$` md5crypt, `$5$` sha256crypt, `$6$` sha512crypt). The `anonymous` user with a `*` password accepts any password |
| `fs` | Backend to use (see [supported backends](#supported-backends)) |
| `params` | Backend parameters, as strings. See each backend's documentation |
| `read_only`, `shared`, `sync_and_delete`, `logging` | Common options, see [fs/README.md](fs/README.md) |

Accesses are evaluated in their declaration order. A hashed password can be generated with
`htpasswd -nbBC 10 "" 'my-password' | cut -d: -f2`, or by enabling `hash_plaintext_passwords`.

#### Local disk

The `os` backend serves a local directory, defined by the `basePath` parameter. Environment variables (`$NAME`) in
the path are replaced by their value.

```json
{
  "user": "test",
  "pass": "test",
  "fs": "os",
  "params": {
    "basePath": "/srv/ftp/$FTP_DIR"
  }
}
```

### Webhook authentication

When `accesses_webhook` is defined, the `accesses` of the config file are ignored: for each login, the user and
password are posted as JSON (`{"user": "...", "pass": "..."}`) to the webhook. It must reply with a `200` status code
and an [access](#accesses) object, any other status code refuses the login.

```json
{
  "accesses_webhook": {
    "url": "https://auth.example.com/ftp",
    "headers": {
      "Authorization": "Bearer my-secret-token"
    },
    "timeout": "5s"
  },
  "accesses": []
}
```

The `timeout` defaults to `30s`.

### TLS

You can generate a self-signed TLS key pair with the following command:

```sh
openssl req -new -newkey rsa:4096 -x509 -sha256 -days 365 -nodes -out cert.pem -keyout key.pem
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
