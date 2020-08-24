package telegram

import (
	"fmt"

	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newChangeDialogTokenDialog() *dialog.Dialog {

	return t.newSelectAlertDialog("changeToken", func(update Update, alert models.Alert) (string, error) {

		updatedAlert, err := t.data.UpdateAlertToken(alert)

		if err != nil {
			return "", err
		}

		return fmt.Sprintf("The token of %s was changed.\n\nYour knew URL is:\n\n%s",
			updatedAlert.Name,
			t.webserver.AlertTriggerURL(updatedAlert, "Hello World"),
		), nil

	})

}
