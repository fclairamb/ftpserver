# Dropbox backend

## Parameters

- `token`: The Dropbox access token to use. If it isn't set, the `DROPBOX_TOKEN` environment variable is used.

## How to use it

### Create an app

1. Create an app in the [Dropbox App Console](https://www.dropbox.com/developers/apps), with the access type you
   need ("App folder" restricts the FTP server to a dedicated folder, "Full Dropbox" gives access to everything).
2. In the "Permissions" tab, enable at least `files.metadata.read`, `files.content.read` and `files.content.write`.
3. In the "Settings" tab, generate an access token.

### Configure the access

```json
{
  "accesses": [
    {
      "user": "dropbox",
      "pass": "dropbox",
      "fs": "dropbox",
      "params": {
        "token": "sl.*****"
      }
    }
  ]
}
```
