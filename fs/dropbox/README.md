# Dropbox backend

## Parameters
Possible parameters are:
- `token`: The non-expiring authentication token to use

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
      "user": "dropbox",
      "pass": "dropbox",
      "fs": "dropbox",
      "params": {
        "google_client_id": "*****.apps.googleusercontent.com",
        "google_client_secret": "*****"
      }
    }
  ]
}
```

