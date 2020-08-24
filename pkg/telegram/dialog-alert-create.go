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

	return dialog.Chain(
		func(i interface{}, ctx dialog.ValueStore) error {
			var err error
			update := i.(Update)

			if update.Update.Message.Command() != "create" {
				return dialog.ErrNoMatch
			}

			// ask for the name of the alert
			response := "Creating a new alert\n\n" +
				"Send me the name of the new alert."

			msg := tgbotapi.NewMessage(update.ChatID, response)
			msg.ParseMode = tgbotapi.ModeMarkdown
			t.bot.Send(msg)

			return err
		},
	).Chain(
		func(i interface{}, ctx dialog.ValueStore) error {
			var err error
			update := i.(Update)

			// get name
			name := strings.TrimSpace(update.Text)

			msg := tgbotapi.NewMessage(update.ChatID, "")

			// check if name matches the defined pattern
			if namePattern.MatchString(name) {
				// store data in context
				ctx.Set("alert", models.Alert{
					Name: name,
				})

				msg.Text = fmt.Sprintf(
					"The name of your new notification is *%s*.\n"+
						"Now give it a description:",
					name)

				msg.ParseMode = tgbotapi.ModeMarkdown
			} else {
				msg.Text = "The profided name does not match the guidelines."
				t.bot.Send(msg)
			}

			err = errInvalidName
			return err
		},
	).Chain(
		func(i interface{}, ctx dialog.ValueStore) error {
			var err error

			// get description
			update := i.(Update)
			description := strings.TrimSpace(update.Text)

			msg := tgbotapi.NewMessage(update.Update.Message.Chat.ID, "")

			// check if matches the description pattern
			if descriptionPattern.MatchString(description) {

				// read data from context
				newAlert, ok := ctx.Value("alert").(models.Alert)
				if ok {
					newAlert.Description = description

					// store new description in context
					ctx.Set("alert", newAlert)

					msg.Text = fmt.Sprintf(
						"Creating new alert:\n\n"+
							"Name: *%q*\n"+
							"Description: *%q*\n\n"+
							"Create: Yes/No",
						newAlert.Name, newAlert.Description)
					msg.ReplyMarkup = yesNoMenuKeyboard
					msg.ParseMode = tgbotapi.ModeMarkdown
				} else {
					err = errContextDataMissing
					msg.Text = err.Error()
				}
			} else {
				err = errInvalidDescription
				msg.Text = err.Error()
			}

			t.bot.Send(msg)
			return err
		},
	).Chain(
		func(i interface{}, ctx dialog.ValueStore) error {
			update := i.(Update)
			var err error

			msg := tgbotapi.NewMessage(update.ChatID, "")

			// check if yes was selected
			if update.Update.Message.Text == "Yes" {
				newAlert, ok := ctx.Value("alert").(models.Alert)
				if ok {
					// create new alert now
					owner := update.User

					a, err := t.data.CreateAlert(
						newAlert.Name,
						newAlert.Description,
						owner,
					)

					if err != nil {
						msg.Text = "Could not create the alert, please try again."
						t.bot.Send(msg)
						return err
					}

					triggerURL := t.webserver.AlertTriggerURL(a, "Hello World")

					msg.Text = fmt.Sprintf(
						"New alert was created\n\n"+
							"To trigger the alarm send a GET request to the follwing URL:\n\n"+
							"[%s](%s)",
						triggerURL, triggerURL)

				} else {
					err = errContextDataMissing
					msg.Text = err.Error()
				}
			} else if update.Update.Message.Text == "No" {
				msg.Text = "Alert is discarded."
				ctx.Set("alert", nil)
			}

			msg.ReplyMarkup = tgbotapi.ReplyKeyboardHide{
				HideKeyboard: true,
			}
			msg.ParseMode = tgbotapi.ModeMarkdown
			t.bot.Send(msg)

			return err
		},
	)
}
