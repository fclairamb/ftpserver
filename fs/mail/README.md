# Mail backend

This is a write-only backend: each uploaded file is sent as an email attachment. Listing directories always returns
an empty result.

## Parameters

- `Host`: The SMTP server.
- `Port` (optional): The SMTP port. Defaults to `25`.
- `SSL` (optional): `true` to connect with SSL/TLS (SMTPS). Defaults to `true` on port `465`, `false` otherwise.
- `StartTLSPolicy` (optional): `MandatoryStartTLS` (default), `OpportunisticStartTLS` or `NoStartTLS`.
- `Username`, `Password` (optional): The SMTP credentials.
- `Localname` (optional): The host name sent in the `HELO` command.
- `From`: The sender address.
- `To`: The recipient address.
- `Subject`: The subject of the emails.
- `MaxSize` (optional): Maximum size of an uploaded file in bytes, `26214400` (25 MiB) by default. Must be positive. Larger
  uploads are refused and no email is sent.
- `Message`: The body of the emails, `%s` is replaced by the file name.

## Configuration

See [example.json](example.json):

```json
{
  "accesses": [
    {
      "user": "test",
      "pass": "test",
      "fs": "mail",
      "params": {
        "From": "alice@example.com",
        "To": "bob@example.com",
        "Subject": "New File",
        "Message": "A new file has been uploaded: %s",
        "Host": "smtp.example.com",
        "Port": "465",
        "Username": "smtpuser",
        "Password": "smtppassword"
      }
    }
  ]
}
```
