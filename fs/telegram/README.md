# FTPServer Telegram connector

[![Stand With Ukraine](https://raw.githubusercontent.com/vshymanskyy/StandWithUkraine/main/banner2-direct.svg)](https://stand-with-ukraine.pp.ua)

## Register bot

Read about telegram bots at https://core.telegram.org/bots/tutorial.

Bots are not allowed to contact users. You need to make the first contact from the user for which you want to set up the bot.

## Quick start

- Create a bot with [@BotFather](https://t.me/BotFather), let's say with username `my_ftp_bot`
- Get bot token from BotFather's response, use it as `Token` in config
- Get bot id by run `curl https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getMe`
- Find `@my_ftp_bot` in telegram and start chat with it
- Send `/start` to bot
- Run `curl https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates` and find your chat id in response, use it as `ChatID` in config


## Config example

Please note about `shared` flag. If it's `true` then bot instance will be shared between all connections.
If it's `false` then each user (or even each ftp connection) will have own bot instance and it can lead to telegram bot flood protection.

`MaxPartSize` (optional) sets the maximum size in bytes for each part when uploading large files. Files exceeding this size are automatically split into multiple parts. Default is `51380224` (49 MB). Telegram's upload limit is 50 MB.

`TempDir` (optional) sets the directory used to store temporary multipart chunks before upload. If omitted, the system temp directory is used.

`RetryAttempts` (optional) sets the number of retry attempts for transient Telegram errors (e.g. rate-limit, timeout). Default is `10`.

`RetryDelay` (optional) sets the delay in milliseconds between retry attempts. Default is `2000` (2 seconds).

`PartUploadDelay` (optional) sets the delay in milliseconds between uploading consecutive parts, to help avoid Telegram rate-limit errors. Default is `500`.

```json
{
  "version": 1,
  "accesses": [
    {
      "fs": "telegram",
      "shared": true,
      "user": "my_ftp_bot",
      "pass": "my_secure_password",
      "params": {
        "Token": "<YOUR_BOT_TOKEN>",
        "ChatID": "<YOUR_CHAT_ID>",
        "MaxPartSize": "51380224",
        "TempDir": "/tmp",
        "RetryAttempts": "10",
        "RetryDelay": "2000",
        "PartUploadDelay": "500"
      }

    }
  ],
  "passive_transfer_port_range": {
    "start": 2122,
    "end": 2130
  }
}
```

## Reassembling multipart files

When a file exceeds `MaxPartSize`, it is uploaded as numbered parts (`filename.part1ofN`, `filename.part2ofN`, …).
Download all parts into the same directory, then reassemble them with one of the commands below.

**Linux / macOS**

```bash
# Reassemble every *.tar.gz split in the current directory
for base in $(ls *.part1of* | sed 's/\.part1of.*//'); do
  ls "${base}.part"*of* | sort -t'f' -k2 -n | xargs cat > "${base}"
  echo "Reassembled: ${base}"
done
```

**Windows (PowerShell)**

```powershell
# Reassemble every *.tar.gz split in the current directory
Get-ChildItem '*.part1of*' | ForEach-Object {
  $base = $_.Name -replace '\.part1of.*', ''
  $out  = $base
  Get-ChildItem "${base}.part*of*" |
    Sort-Object { [int]($_.Name -replace '.*\.part(\d+)of.*','$1') } |
    ForEach-Object { Get-Content $_.FullName -AsByteStream } |
    Set-Content -AsByteStream $out
  Write-Host "Reassembled: $out"
}
```
