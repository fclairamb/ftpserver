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
- `chmod_acl` (optional): Set to `"true"` to let `SITE CHMOD` change the object's canned ACL (see below). Off by default: `SITE CHMOD` is then refused with an error.

### `SITE CHMOD` and object ACLs

By default, `SITE CHMOD` (and any `Chmod` call) is refused on an s3 access with the error `chmod is not supported on s3`,
which the FTP client receives as a 550 reply. The backend isn't called.

With `"chmod_acl": "true"`, the mode is mapped to a canned ACL and applied with `PutObjectAcl`:

- "other" read and write bits set (`666`, `777`...): `public-read-write`
- "other" read bit set (`644`, `755`...): `public-read`
- otherwise: `private` (which replaces the object's ACL and drops any other grant on it)

> **Warning:** `chmod_acl` lets every user of the access make any object of its prefix public, including files
> uploaded by other users, as soon as the bucket accepts ACLs, its public access isn't blocked and the server's
> credentials allow `s3:PutObjectAcl`. Some FTP clients send `SITE CHMOD 644` after each upload, which then publishes
> every uploaded file. Only enable it if you rely on `SITE CHMOD` to publish files on purpose.
>
> Versions 0.9.0 to 0.16.7 always behaved as if `chmod_acl` was set. If your bucket accepts ACLs, check that no object
> was made public through `SITE CHMOD`.

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
- Ownership is a no-op since S3 does not support it. Permissions are refused unless `chmod_acl` is set (see above).
- Seeking within files is not supported for writes.
