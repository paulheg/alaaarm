package telegram

import (
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newDeleteDialog() *dialog.Dialog {
	return t.newSelectDialog("delete", func(alert models.Alert) (string, error) {

		err := t.data.DeleteAlert(alert)
		if err != nil {
			return "", err
		}

		return "Alert was deleted.", nil
	})
}
