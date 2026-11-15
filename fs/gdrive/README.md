# Google Drive backend

## Parameters

- `google_client_id`: The Google client ID. If it isn't set, the `GOOGLE_CLIENT_ID` environment variable is used.
- `google_client_secret`: The Google client secret, it shall never be shared. If it isn't set, the
  `GOOGLE_CLIENT_SECRET` environment variable is used.
- `token_file` (optional): The path of the token file that will be created. Defaults to `gdrive_token_<user>.json`.
- `base_path` (optional): The Google Drive folder to use as the root directory. Environment variables (`$NAME`) are
  replaced by their value.

_Note:_ The token file is updated regularly (every day). The server must be able to write the token file after its
initial creation.

## How to use it

### Create an app

1. In the [Google Cloud console](https://console.cloud.google.com/), create a project and enable the Google Drive API.
2. Configure the OAuth consent screen.
3. Create an "OAuth client ID" credential of the "Desktop app" type, and note its client ID and client secret.

### Configure the access

```json
{
  "accesses": [
    {
      "user": "gdrive",
      "pass": "gdrive",
      "fs": "gdrive",
      "params": {
        "google_client_id": "*****.apps.googleusercontent.com",
        "google_client_secret": "*****",
        "base_path": "ftp"
      }
    }
  ]
}
```

### Connect for the first time

The first time you log in with the "gdrive" user, the server logs an authorization link and waits. Open it in a
browser and grant access: Google then redirects the browser to a temporary `http://127.0.0.1:<port>/` page served by
the server, which captures the authorization code (Google blocked the old copy-paste "out-of-band" flow). The server
then creates the token file, and won't ask you again.

If the browser doesn't run on the server's machine (remote server, docker container), the redirect fails with a
"can't connect" page. Copy the URL from the address bar and open it from the server instead, before the login times
out on the client side:

```sh
curl 'http://127.0.0.1:<port>/?state=...&code=...'
```

You can also authorize on your own machine and copy the token file to the server.
