package telegram

import (
	"fmt"

	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newChangeDialogTokenDialog() *dialog.Dialog {

	return t.newSelectDialog("changeToken", func(alert models.Alert) (string, error) {

		updatedAlert, err := t.data.UpdateAlertToken(alert)

		if err != nil {
			return "", err
		}

		return fmt.Sprintf("The token of %s was changed.\n\nYour knew URL is:\n\n%s",
			updatedAlert.Name,
			"URL",
		), nil

	})

}
