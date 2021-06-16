# Golang FTP Server

![Build](https://github.com/fclairamb/ftpserver/workflows/Build/badge.svg)
![Cross Build](https://github.com/fclairamb/ftpserver/workflows/Cross%20Build/badge.svg)
![Docker Image](https://github.com/fclairamb/ftpserver/workflows/Docker%20Image/badge.svg)
![Docker test](https://github.com/fclairamb/ftpserver/workflows/Docker%20test/badge.svg)
![Docker downloads](https://img.shields.io/docker/pulls/fclairamb/ftpserver)
[![Go Report Card](https://goreportcard.com/badge/fclairamb/ftpserver)](https://goreportcard.com/report/fclairamb/ftpserver)
[![GoDoc](https://godoc.org/github.com/fclairamb/ftpserver?status.svg)](https://godoc.org/github.com/fclairamb/ftpserver/server)

This FTP server is a gateway between old-school FTP devices and modern cloud based file systems, using the
[afero](https://github.com/spf13/afero)'s Fs interface and acting as a reference implementation of the [ftpserverlib](https://github.com/fclairamb/ftpserverlib).

At the current stage, supported backend are:
- Local disk
- [S3](https://aws.amazon.com/s3/) through [afero-s3](https://github.com/fclairamb/afero-s3)
- [Dropbox](https://www.dropbox.com/) (see [doc](https://github.com/fclairamb/ftpserver/tree/master/fs/dropbox)) through [afero-dropbox](https://github.com/fclairamb/afero-dropbox)
- [Google Drive](https://developers.google.com/drive) (see [doc](https://github.com/fclairamb/ftpserver/tree/master/fs/gdrive)) through [afero-gdrive](https://github.com/fclairamb/afero-gdrive)
- [SFTP](https://en.wikipedia.org/wiki/SSH_File_Transfer_Protocol) through [afero's sftpfs](https://github.com/spf13/afero/)
- Email through [go-mail](https://github.com/go-mail/mail) thanks to [@x-way](https://github.com/x-way)

## Current status of the project

### Features

#### FTP protocol

These features are brought by [ftpserverlib](https://github.com/fclairamb/ftpserverlib) itself:

 * Uploading and downloading files
 * Directory listing (LIST + MLST)
 * File and directory deletion and renaming
 * TLS support (AUTH + PROT)
 * File download/upload resume support (REST)
 * Complete driver for all the above features
 * Passive socket connections (EPSV and PASV commands)
 * Active socket connections (PORT command)
 * Small memory footprint
 * Only relies on the standard library except for:
   * [go-kit log](https://github.com/go-kit/kit/tree/master/log) for logging
   * [afero](https://github.com/spf13/afero) for generic file systems handling
 * Supported extensions:
   * [AUTH](https://tools.ietf.org/html/rfc2228#page-6) - Control session protection
   * [AUTH TLS](https://tools.ietf.org/html/rfc4217#section-4.1) - TLS session
   * [PROT](https://tools.ietf.org/html/rfc2228#page-8) - Transfer protection
   * [MDTM](https://tools.ietf.org/html/rfc3659#page-8) - File Modification Time
   * [SIZE](https://tools.ietf.org/html/rfc3659#page-11) - Size of a file
   * [REST](https://tools.ietf.org/html/rfc3659#page-13) - Restart of interrupted transfer
   * [MLST](https://tools.ietf.org/html/rfc3659#page-23) - Simple file listing for machine processing
   * [MLSD](https://tools.ietf.org/html/rfc3659#page-23) - Directory listing for machine processing

## Getting started

### Get it
#### Golang

```bash
go get -u github.com/fclairamb/ftpserver
```

### Config file
If you don't create a `ftpserver.json` file, it will be created for you.

Here is a sample config file:

```json
{
   "version": 1,
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
         "user": "test",
         "pass": "test",
         "fs": "os",
         "params": {
            "basePath": "/tmp"
         }
      },
      {
         "user": "dropbox",
         "pass": "dropbox",
         "fs": "dropbox",
         "params": {
            "token": "..."
         }
      },
      {
         "user": "gdrive",
         "pass": "gdrive",
         "fs": "gdrive",
         "params": {
            "google_client_id": "***.apps.googleusercontent.com",
            "google_client_secret": "****"
         }
      },
      {
         
         "user": "s3",
         "pass": "s3",
         "fs": "s3",
         "params": {
            "endpoint": "https://s3.amazonaws.com",
            "region": "eu-west-1",
            "bucket": "my-bucket",
            "access_key_id": "AKIA....",
            "secret_access_key": "IDxd....",
            "disable_ssl": "false",
            "path_style": "false"
         }
      },
      {
         "user": "sftp",
         "pass": "sftp",
         "fs": "sftp",
         "params": {
            "username": "user",
            "password": "password",
            "hostname": "192.168.168.11:22"
         }
      }
   ],
   "passive_transfer_port_range": {
      "start": 2122,
      "end": 2130
   }
}
```

### With local binary
You can build the binary and use it directly:

```sh
# Get and install the server
go get github.com/fclairamb/ftpserver

ftpserver &

# Download some file
[ -f kitty.jpg ] || (curl -o kitty.jpg.tmp https://placekitten.com/2048/2048 && mv kitty.jpg.tmp kitty.jpg)

# Upload it to the server
curl -v -T kitty.jpg ftp://test:test@localhost:2121/

# Download it back
curl ftp://test:test@localhost:2121/kitty.jpg -o kitty2.jpg

# Compare it
diff kitty.jpg kitty2.jpg
```

### With docker
There's also a containerized version of the server (15MB, based on alpine).

```sh
# Starting the sample FTP server
docker run --rm -d -p 2121-2130:2121-2130 -v ./ftpserver/files:/tmp -v ./ftpserver:/app fclairamb/ftpserver

# Download some file
[ -f kitty.jpg ] || (curl -o kitty.jpg.tmp https://placekitten.com/2048/2048 && mv kitty.jpg.tmp kitty.jpg)

# Upload it
curl -v -T kitty.jpg ftp://test:test@localhost:2121/

# Download it back
curl ftp://test:test@localhost:2121/kitty.jpg -o kitty2.jpg

# Compare it
diff kitty.jpg kitty2.jpg
```

### With Docker Compose
```yml
# docker-compose.yml

version: '3.3'
services:
  ftpserver:
    ports:
      - '2121-2130:2121-2130'
    volumes:
      - ./ftpserver/files:/tmp
      - ./ftpserver:/app
    image: fclairamb/ftpserver
```

```sh
docker-compose up -d
```
