package telegram

import (
	"regexp"
	"strings"

	"github.com/kyokomi/emoji"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newCreateAlertDialog() *dialog.Dialog {

	namePattern := regexp.MustCompile("^.{5,}$")
	descriptionPattern := regexp.MustCompile("^.{10,}$")

	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		// ask for the name of the alert
		msg := tgbotapi.NewMessage(u.ChatID, "")
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.Text = emoji.Sprint(`:bell: Creating a new alert
		
		Send me the name of the new alert.`)
		_, err := t.bot.Send(msg)
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Success, nil
	})).Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		msg := tgbotapi.NewMessage(u.ChatID, "")

		name := strings.TrimSpace(u.Text)

		// check if name matches the defined pattern
		if !namePattern.MatchString(name) {
			msg.Text = emoji.Sprint(":warning: The alert name has to be at least 5 characters long.\nPlease try again.")
			_, err := t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Retry, nil
		}

		// store data in context
		ctx.Set("alert", models.Alert{
			Name: name,
		})

		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.Text = emoji.Sprintf(`The name of your new notification is *%s*.

Now give it a description.`, name)

		_, err := t.bot.Send(msg)
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Success, nil
	})).Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		msg := tgbotapi.NewMessage(u.ChatID, "")

		description := strings.TrimSpace(u.Text)

		// check if matches the description pattern
		if !descriptionPattern.MatchString(description) {
			msg.Text = emoji.Sprint(":warning: The description has to be at least 10 characters long.\nPlease try again.")
			_, err := t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}
			return dialog.Retry, nil
		}

		// read data from context
		newAlert, ok := ctx.Value("alert").(models.Alert)

		// fatal error, should not happen
		if !ok {
			return dialog.Reset, errContextDataMissing
		}

		newAlert.Description = description

		// store new description in context
		ctx.Set("alert", newAlert)

		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.Text = emoji.Sprintf(`:bell: Creating new alert
*Name:* %s
*Description:* %s

Do you want to create the alert?`, newAlert.Name, newAlert.Description)

		_, err := t.bot.Send(msg)
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Next, nil
	})).Append(t.newYesNoDialog(
		// On Yes
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			msg := tgbotapi.NewMessage(u.ChatID, "")

			contextAlert, ok := ctx.Value("alert").(models.Alert)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			alert := models.NewAlert(contextAlert.Name, contextAlert.Description, u.User)

			a, err := t.repository.CreateAlert(*alert)

			if err != nil {
				return dialog.Reset, err
			}

			triggerURL := t.webserver.AlertTriggerURL(a, "Hello World")

			msg.ParseMode = tgbotapi.ModeMarkdown
			msg.Text = emoji.Sprintf(`:check_mark_button: New alert :bell: was created

To trigger the alarm send a GET request to the follwing URL:
[%s](%s)`, triggerURL, triggerURL)

			_, err = t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		},
		// On no
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			msg := tgbotapi.NewMessage(u.ChatID, "")
			msg.Text = emoji.Sprint(":cross_mark: Alert is discarded.")
			_, err := t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Reset, nil
		},
	))
}
