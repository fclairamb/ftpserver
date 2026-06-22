# AWS S3 filesystem

This directory contains the AWS S3 implementation for the FTP server using the official AWS SDK v2 and [`afero-s3`](https://github.com/fclairamb/afero-s3).

## Configuration

```json
{
  "version": 1,
  "accesses": [
    {
      "user": "test",
      "pass": "test",
      "fs": "s3",
      "params": {
        "bucket": "my-bucket",
        "region": "us-east-1",
        "access_key_id": "AKIAIOSFODNN7EXAMPLE",
        "secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
      }
    }
  ]
}
```

### Parameters

- `bucket` (required): S3 bucket name.
- `region` (optional): AWS region (e.g. `us-east-1`). Defaults to the environment or SDK chain.
- `access_key_id` (optional): AWS access key ID. If omitted, the default credential chain is used (env vars, IAM role, etc.).
- `secret_access_key` (optional): AWS secret access key. Required when `access_key_id` is set.
- `endpoint` (optional): Custom endpoint URL for S3-compatible storage (MinIO, Ceph, LocalStack, etc.).
- `disable_ssl` (optional): Set to `"true"` to use HTTP instead of HTTPS for the endpoint.
- `path_style` (optional): Set to `"true"` to use path-style addressing (`host/bucket/key`) instead of virtual-hosted style (`bucket.host/key`). Required for most S3-compatible backends.
- `basePath` (optional): Key prefix within the bucket to use as the root directory. Useful for isolating users to a subdirectory. Leading and trailing slashes are stripped automatically.

### Example configurations

#### Standard AWS S3 with IAM role (no explicit credentials):
```json
{
  "fs": "s3",
  "params": {
    "bucket": "my-ftp-bucket",
    "region": "eu-west-1"
  }
}
```

#### S3-compatible backend (e.g. MinIO):
```json
{
  "fs": "s3",
  "params": {
    "bucket": "my-bucket",
    "endpoint": "http://minio:9000",
    "path_style": "true",
    "access_key_id": "minioadmin",
    "secret_access_key": "minioadmin"
  }
}
```

#### Isolated user directory via `basePath`:
```json
{
  "fs": "s3",
  "params": {
    "bucket": "shared-bucket",
    "region": "us-east-1",
    "basePath": "users/alice"
  }
}
```

The FTP root `/` maps to the S3 prefix `users/alice/`. Files uploaded to `/report.csv` are stored as `users/alice/report.csv` in the bucket.

### Notes

- S3 has no real directories; directory operations are emulated using key prefixes.
- File permissions and ownership are no-ops since S3 does not support them.
- Seeking within files is not supported for writes.
