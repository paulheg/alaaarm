package telegram

import (
	"fmt"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newCreateAlertDialog() *dialog.Dialog {

	namePattern := regexp.MustCompile("^.{5,}$")
	descriptionPattern := regexp.MustCompile("^.{10,}$")

	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
		if u.Update.Message.Command() != "create" {
			return dialog.NoMatch, nil
		}

		// ask for the name of the alert
		response := "Creating a new alert\n\n" +
			"Send me the name of the new alert."

		msg := tgbotapi.NewMessage(u.ChatID, response)
		msg.ParseMode = tgbotapi.ModeMarkdown
		t.bot.Send(msg)

		return dialog.Success, nil
	})).Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		msg := tgbotapi.NewMessage(u.ChatID, "")

		name := strings.TrimSpace(u.Text)

		// check if name matches the defined pattern
		if !namePattern.MatchString(name) {
			msg.Text = "The alert name has to be at least 5 characters long.\nPlease try again."
			t.bot.Send(msg)
			return dialog.Retry, nil
		}

		// store data in context
		ctx.Set("alert", models.Alert{
			Name: name,
		})

		msg.Text = fmt.Sprintf(
			"The name of your new notification is *%s*.\n"+
				"Now give it a description:",
			name)

		msg.ParseMode = tgbotapi.ModeMarkdown
		t.bot.Send(msg)
		return dialog.Success, nil
	})).Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		msg := tgbotapi.NewMessage(u.ChatID, "")

		description := strings.TrimSpace(u.Text)

		// check if matches the description pattern
		if !descriptionPattern.MatchString(description) {
			msg.Text = "The description has to be at least 10 characters long. Please try again."
			t.bot.Send(msg)
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

		msg.Text = fmt.Sprintf(
			"Creating new alert:\n\n"+
				"Name: *%q*\n"+
				"Description: *%q*\n\n"+
				"Do you want to create the alert?",
			newAlert.Name, newAlert.Description)
		msg.ParseMode = tgbotapi.ModeMarkdown

		t.bot.Send(msg)
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

			msg.Text = fmt.Sprintf(
				"New alert was created\n\n"+
					"To trigger the alarm send a GET request to the follwing URL:\n\n"+
					"[%s](%s)",
				triggerURL, triggerURL)

			msg.ParseMode = tgbotapi.ModeMarkdown

			t.bot.Send(msg)
			return dialog.Success, nil
		},
		// On no
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
			msg := tgbotapi.NewMessage(u.ChatID, "Alert is discarded.")
			t.bot.Send(msg)
			return dialog.Reset, nil
		},
	))
}
