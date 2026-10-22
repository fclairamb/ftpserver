# Google Cloud Storage (GCS) filesystem

This backend gives access to a Google Cloud Storage bucket, using afero's
[gcsfs](https://github.com/spf13/afero/tree/master/gcsfs) package.

## Configuration

The GCS filesystem supports the following configuration parameters:

```json
{
  "version": 1,
  "accesses": [
    {
      "user": "test",
      "pass": "test",
      "fs": "gcs",
      "params": {
        "bucket": "your-gcs-bucket-name",
        "key_file": "/path/to/service-account-key.json"
      }
    }
  ]
}
```

### Parameters

- `bucket` (required): The GCS bucket name to use. It is the root directory of the FTP access.
- `key_file` (optional): Path to a service account key JSON file for authentication. If not provided, the default
  credentials are used (environment variables, metadata service, etc.)
- `project_id`: Not used (the bucket name is enough to access it), it is accepted for backward compatibility.

### Authentication

The GCS filesystem supports multiple authentication methods:

1. **Service Account Key File**: Provide the `key_file` parameter pointing to a JSON service account key file
2. **Default Credentials**: If `key_file` is not provided, it will use Google Cloud's default credential chain:
   - Environment variable `GOOGLE_APPLICATION_CREDENTIALS`
   - User credentials from `gcloud auth application-default login`
   - Service account attached to the resource (when running on Google Cloud)

### Example configurations

#### With service account key file:
```json
{
  "fs": "gcs",
  "params": {
    "bucket": "my-ftp-bucket",
    "key_file": "/etc/gcs-key.json"
  }
}
```

#### With default credentials:
```json
{
  "fs": "gcs",
  "params": {
    "bucket": "my-ftp-bucket"
  }
}
```

### Notes

- GCS doesn't have real directories, so directory operations are emulated
- File permissions and ownership operations are no-ops since GCS doesn't support them
- Seeking to end of file and some advanced file operations are not supported due to GCS limitations
