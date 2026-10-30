# SFTP backend

This backend connects to an upstream SSH/SFTP server.

## Parameters

- `hostname`: The upstream server address, with its port (e.g. `192.168.1.10:22`).
- `username`: The SSH user.
- `password`: The SSH password.
- One of the host key verification parameters below.

## Host key verification

The upstream server's host key is verified to protect against man-in-the-middle attacks. You must configure exactly
one of the following parameters:

- `known_hosts`: path to an OpenSSH `known_hosts` file containing the upstream server's key
  (e.g. generated with `ssh-keyscan -H host >> known_hosts`).
- `host_key`: the upstream server's public key pinned inline, in `authorized_keys` / `known_hosts` line format
  (e.g. `ssh-ed25519 AAAA...`).
- `insecure_ignore_host_key`: set to `"true"` to disable host key verification entirely. This is **unsafe** (it
  exposes the connection to man-in-the-middle attacks) and should only be used in trusted networks or for testing.

If none of these is set, the connection is refused.

## Configuration

```json
{
  "accesses": [
    {
      "user": "sftp",
      "pass": "sftp",
      "fs": "sftp",
      "shared": true,
      "params": {
        "hostname": "192.168.1.10:22",
        "username": "user",
        "password": "password",
        "known_hosts": "/home/ftpserver/.ssh/known_hosts"
      }
    }
  ]
}
```

With `shared` enabled, a single SSH connection is used for all the FTP sessions of this access.
