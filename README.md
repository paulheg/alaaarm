# 👨‍🚒 ALAAARM

Alaaarm is essentially a Telegram bot with a web interface, trough which you can trigger notifications.

The idea of the project arose from the desire to send notifications to a smartphone without much effort, from different sources with little integration effort. Users are enabled to create custom alerts trough the Telegram bot. Each alert is triggered trough the public web interface using a token as an unique identifier.

## Example

With this command & URL you send the message "Hello World" to all subscribed users of the `{token}` alert.

````bash
curl -X GET "http://localhost/api/v1/alert/{token}/trigger?m=Hello+World"
````

With this command & URL you send the image `test.jpg` with the caption "I showed you my code please respond".

````bash
curl -X POST "http://localhost/api/v1/alert/{token}/trigger?m=I+showed+you+my+code+please+respond" -F "image=@test.jpg"
````

> replace localhost with the actual domain *alaaarm* is hosted on

## Q&A

> Can multiple users receive notifications from the same alert?

To enable multiple users to receive notifications, the owner/creator of the alert has to generate an invitation link using the `/invite` command. This link can be shared with other Telegram users to subscribe them to the selected alert.

> HELP!! I am getting spammed, the trigger URL / token got leaked.

Don't worry, you can change the token of an alert with the command `\change_alert_token`, which leads to the old URL getting invalid and therefore you wont receive any messages from it.

## Commands

| Command               | Description                                       |
| --------------------- | ------------------------------------------------- |
| `\create`             | create new alert                                  |
| `\delete`             | delete an alert you created                       |
| `\info`               | get info about your created and subscribed alerts |
| `\alert_info`         | get detailed info about an alert                  |
| `\unsubscribe`        | unsubscribe from an alert you were invited to     |
| `\change_alert_token` | change the alert token                            |
| `\invite`             | create an invitation link for your alert          |
| `\delete_invite`      | delete a previously created invite                |
| `\exit`               | exit current action and reset the dialog          |
