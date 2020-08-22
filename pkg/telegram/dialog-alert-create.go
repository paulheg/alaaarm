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

			// answer with creating options
			update := i.(tgbotapi.Update)

			if update.Message.IsCommand() && update.Message.Command() == "create" {
				response := "Creating a new alert\n\n" +
					"Send me the name of the new alert."

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
				msg.ParseMode = tgbotapi.ModeMarkdown
				t.bot.Send(msg)
			} else {
				err = dialog.ErrNoMatch
			}

			return err
		},
	).Chain(
		func(i interface{}, ctx dialog.ValueStore) error {
			var err error

			// get name
			update := i.(tgbotapi.Update)
			name := strings.TrimSpace(update.Message.Text)

			// check if name matches the defined pattern
			if namePattern.MatchString(name) {
				// store data in context
				ctx.Set("alert", models.Alert{
					Name: name,
				})

				response := fmt.Sprintf(
					"The name of your new notification is *%s*.\n"+
						"Now give it a description:",
					name)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
				msg.ParseMode = tgbotapi.ModeMarkdown
				t.bot.Send(msg)
			} else {
				err = errInvalidName
			}

			return err
		},
	).Chain(
		func(i interface{}, ctx dialog.ValueStore) error {
			var err error

			// get description
			update := i.(tgbotapi.Update)
			description := strings.TrimSpace(update.Message.Text)

			// check if matches the description pattern
			if descriptionPattern.MatchString(description) {

				// read data from context
				newAlert, ok := ctx.Value("alert").(models.Alert)
				if ok {
					newAlert.Description = description

					// store new description in context
					ctx.Set("alert", newAlert)

					response := fmt.Sprintf(
						"Creating new alert:\n\n"+
							"Name: *%q*\n"+
							"Description: *%q*\n\n"+
							"Create: Yes/No",
						newAlert.Name, newAlert.Description)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
					msg.ReplyMarkup = yesNoMenuKeyboard
					msg.ParseMode = tgbotapi.ModeMarkdown
					t.bot.Send(msg)

				} else {
					err = errContextDataMissing
				}
			} else {
				err = errInvalidDescription
			}

			return err
		},
	).Chain(
		func(i interface{}, ctx dialog.ValueStore) error {
			update := i.(tgbotapi.Update)
			var response string
			var err error

			// check if yes was selected
			if update.Message.Text == "Yes" {
				newAlert, ok := ctx.Value("alert").(models.Alert)
				if ok {
					// create new alert now

					// get user from telegramID or create one if does not exist
					var owner models.User
					exists, owner, err := t.data.GetUserTelegram(update.Message.Chat.ID)
					if !exists {
						owner, err = t.data.CreateUser(models.User{
							Username:   update.Message.Chat.UserName,
							TelegramID: update.Message.Chat.ID,
						})
					}

					if err != nil {
						// something went wrong in the database
						response = "Could not find or create a user, please try again."

					} else {
						a, err := t.data.CreateAlert(
							newAlert.Name,
							newAlert.Description,
							owner)

						if err == nil {
							triggerURL := fmt.Sprintf("http://st1cker.com/api/v1/alert/%s/trigger?m=HelloWorld", a.Token)

							response = fmt.Sprintf(
								"New alert was created\n\n"+
									"To trigger the alarm send a GET request to the follwing URL:\n\n"+
									"[%s](%s)",
								triggerURL, triggerURL)
						} else {
							// something went wrong in the database
							response = "Could not create the alert, please try again."
						}
					}

				} else {
					err = errContextDataMissing
				}

			} else if update.Message.Text == "No" {
				response = "Alert is discarded."
				// cancel all actions
			}

			if len(response) > 0 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
				msg.ReplyMarkup = tgbotapi.ReplyKeyboardHide{
					HideKeyboard: true,
				}
				msg.ParseMode = tgbotapi.ModeMarkdown
				t.bot.Send(msg)
			}

			return err
		},
	)
}
