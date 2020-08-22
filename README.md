# 👨‍🚒 ALAAARM

## What is *ALAAARM*

ALAAARM is a Platform allowing users to create notification hooks for Telegram.

When creating a hook, the bot will create a custom URL, that when creating a request to, will notify all registered users.

```text
[GET]
\api\v1\{token}\trigger?m={text message}

[POST]
\api\v1\{token}\trigger?m={text message}
+ Attachments (Images)
```

As the owner of a hook you can decide if it is public or private and can authorize users for your private hook.

All configuration happens over the bot.

## Commands

- create
- configure
- delete
