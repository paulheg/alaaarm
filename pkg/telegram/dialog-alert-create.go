package telegram

import (
	"regexp"
	"strings"

	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newCreateAlertDialog() *dialog.Dialog {

	namePattern := regexp.MustCompile("^.{5,}$")
	descriptionPattern := regexp.MustCompile("^.{10,}$")

	return dialog.Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		// ask for the name of the alert
		_, err := t.bot.Send(t.escapedHTMLLookup(u.ChatID, "create_new_alert"))
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Success, nil
	})).Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		name := strings.TrimSpace(u.Text)

		// check if name matches the defined pattern
		if !namePattern.MatchString(name) {

			_, err := t.bot.Send(t.escapedHTMLLookup(u.ChatID, "alert_name_too_short"))
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Retry, nil
		}

		// store data in context
		ctx.Set(ALERT_SELECTION_CONTEXT_KEY, models.Alert{
			Name: name,
		})

		_, err := t.bot.Send(t.escapedHTMLLookup(u.ChatID, "alert_needs_description", name))
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Success, nil
	})).Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

		description := strings.TrimSpace(u.Text)

		// check if matches the description pattern
		if !descriptionPattern.MatchString(description) {

			_, err := t.bot.Send(t.escapedHTMLLookup(u.ChatID, "alert_description_too_short"))
			if err != nil {
				return dialog.Reset, err
			}
			return dialog.Retry, nil
		}

		// read data from context
		newAlert, ok := ctx.Value(ALERT_SELECTION_CONTEXT_KEY).(models.Alert)

		// fatal error, should not happen
		if !ok {
			return dialog.Reset, errContextDataMissing
		}

		newAlert.Description = description

		// store new description in context
		ctx.Set(ALERT_SELECTION_CONTEXT_KEY, newAlert)

		_, err := t.bot.Send(t.escapedHTMLLookup(u.ChatID, "alert_finished", newAlert.Name, newAlert.Description))
		if err != nil {
			return dialog.Reset, err
		}

		return dialog.Next, nil
	})).Append(t.newYesNoDialog(
		// On Yes
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			contextAlert, ok := ctx.Value(ALERT_SELECTION_CONTEXT_KEY).(models.Alert)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			alert := models.NewAlert(contextAlert.Name, contextAlert.Description, u.User)

			a, err := t.repository.CreateAlert(*alert)

			if err != nil {
				return dialog.Reset, err
			}

			triggerURL := t.webserver.AlertTriggerURL(a, "Hello World")

			_, err = t.bot.Send(t.escapedHTMLLookup(u.ChatID, "alert_created", triggerURL, triggerURL))
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		},
		// On no
		func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			_, err := t.bot.Send(t.escapedHTMLLookup(u.ChatID, "alert_discarded"))
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Reset, nil
		},
	))
}
