# Google Drive backend

## Parameters
Possible parameters are:
- `google_client_id`: The Google Client ID
- `google_client_secret`: The Google Client Secret, it shall never be shared
- `token_file`: The path of the token file that will be created
- `base_path`: The base path of the Google Drive folder

_Note:_ The token file will be updated regularly (every day). It's important that the server can write the token file 
after its initial creation.

## How to use it
### Create an app
You can follow 
[these instructions](https://medium.com/swlh/google-drive-api-with-python-part-i-set-up-credentials-1f729cb0372b).

### Use the created client credentials
You can either define the `google_client_id` and `google_client_secret` access params or specify the `GOOGLE_CLIENT_ID`
and `GOOGLE_CLIENT_SECRET` environment variables.

You should have a config file looking like this one:
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
        "basePath": "/ftp"
      }
    }
  ]
}
```

### Connect for the first time
You should run the ftpserver in interactive mode. 
The first time you will login with the "gdrive" user, the server will ask you to open a link, wich
will authorize the app you just created and return a code. This code will allow the server to create
a token file (named `gdrive_token_$username.json` by default).

