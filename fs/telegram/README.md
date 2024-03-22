# FTPServer Telegram connector

[![Stand With Ukraine](https://raw.githubusercontent.com/vshymanskyy/StandWithUkraine/main/banner2-direct.svg)](https://stand-with-ukraine.pp.ua)

## Register bot

Read about telegram bots at https://core.telegram.org/bots/tutorial.

Bots are not allowed to contact users. You need to make the first contact from the user for which you want to set up the bot.

### Quick start

- Create a bot with [@BotFather](https://t.me/BotFather), let's say with username `my_ftp_bot`
- Get bot token from BotFather's response, use it as `Token` in config
- Get bot id by run `curl https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getMe`
- Find `@my_ftp_bot` in telegram and start chat with it
- Send `/start` to bot
- Run `curl https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates` and find your chat id in response, use it as `ChatID` in config


## Config example

Please note about `shared` flag. If it's `true` then bot instance will be shared between all connections.
If it's `false` then each user (or even each ftp connection) will have own bot instance and it can lead to telegram bot flood protection.

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
        "ChatID": "<YOUR_CHAT_ID>"
      }

    }
  ],
  "passive_transfer_port_range": {
    "start": 2122,
    "end": 2130
  }
}
```
