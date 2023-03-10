# 👨‍🚒 ALAAARM

Alaaarm is essentially a Telegram bot with a minimal web API, trough which you can trigger notifications.

With it you can send notifications to a smartphone without much setup, from different sources with little integration effort.
Users are enabled to create custom alerts trough the Telegram bot.
Each alert is triggered trough the public web interface using a token as an unique identifier.

Give it a try on: https://t.me/AlaarmAlaaarmBot

## Example

With this command & URL you send the message "Hello World" to all subscribed users of the `{token}` alert.

````bash
curl -X GET "http://localhost/api/v1/alert/{token}/trigger?m=Hello+World"
````

With this command & URL you send the image `test.jpg` with the caption "I sent you my code please respond".

````bash
curl -X POST "http://localhost/api/v1/alert/{token}/trigger?m=I+sent+you+my+code+please+respond" -F "file=@test.jpg"
````

> replace localhost with the actual domain *alaaarm* is hosted on

## Q&A

> Can multiple users receive notifications from the same alert?

There are two options to let multiple users receive the same alert:

1. Add the bot to an existing Telegram Group by generating a group invite with the `/invite` command. You have to be an admin of the group to do this.
2. Create a private invite link with `/invite` and share it with other users.

> HELP!! I am getting spammed, the trigger URL / token got leaked.

Don't worry, you can change the token of an alert with the command `/change_alert_token`, which leads to the old URL getting invalid and therefore you wont receive any messages from it.

> What does the number next to 🗿 mean?

This is the number of subscribed channels / users to your alert.

## Commands

| Command               | Description                                       |
| --------------------- | ------------------------------------------------- |
| `/start`              | start talking to the bot                          |
| `/create`             | create new alert                                  |
| `/delete`             | delete an alert you created                       |
| `/info`               | get info about your created and subscribed alerts |
| `/alert_info`         | get detailed info about an alert                  |
| `/change_alert_token` | change the alert token                            |
| `/invite`             | create an invitation link for your alert          |
| `/delete_invite`      | delete a previously created invite                |
| `/mute`               | Don't get notified from your own alert            |
| `/unsubscribe`        | unsubscribe from an alert you were invited to     |
| `/exit`               | exit current action and reset the dialog          |
