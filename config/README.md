# Sample config files

The config file format is described in the [main README](../README.md#configuration) and by the
[JSON schema](../config-schema.json).

When the server starts without `-conf` and no `ftpserver.json` exists, it creates one with a `test` user serving
`/tmp`. That user gets a random password, printed once in the logs (`docker logs <container>` for the docker
images) and stored bcrypt-hashed in the file. Set your own users and passwords in the file afterwards.

- [behind_nat.json](samples/behind_nat.json): FTP server behind a NAT gateway, advertising its public IP address
  for passive transfers.
- [hashed_passwords.json](samples/hashed_passwords.json): plain-text, bcrypt and sha512crypt passwords (the
  password of each user is its user name).
- [log_file.json](samples/log_file.json): logging to a file.

Each backend also has its own sample in its [documentation](../fs).
