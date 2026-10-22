# Keycloak backend

This backend serves local directories to users authenticated by a [Keycloak](https://www.keycloak.org/) server.
Each user gets their own `<base_path>/<user name>` directory.

## Parameters

- `keycloak_url`: The URL of the Keycloak server.
- `keycloak_realm`: The realm of the users.
- `keycloak_client_id`: The client ID. If it isn't set, the `KEYCLOAK_CLIENT_ID` environment variable is used.
- `keycloak_client_secret`: The client secret. If it isn't set, the `KEYCLOAK_CLIENT_SECRET` environment variable is
  used.
- `base_path`: The local directory containing the users' directories. Environment variables (`$NAME`) are replaced by
  their value.

The client must allow the "Direct Access Grants" (resource owner password credentials) flow, which is how the
credentials sent by the FTP clients are checked.

## Configuration

A keycloak access doesn't have any `user` or `pass`, it accepts any user that Keycloak authenticates. As accesses are
evaluated in their declaration order, it should be declared last.

```json
{
  "accesses": [
    {
      "fs": "keycloak",
      "params": {
        "keycloak_url": "http://localhost:8080",
        "keycloak_realm": "master",
        "keycloak_client_id": "ftpserver",
        "keycloak_client_secret": "*****",
        "base_path": "/srv/ftp"
      }
    }
  ]
}
```
