package telegram

import (
	"fmt"

	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newAlertUnsubscribeDialog() *dialog.Dialog {

	return t.newSelectAlertDialog("unsubscribe", func(update Update, alert models.Alert) (string, error) {

		if update.User.ID == alert.OwnerID {
			return "You cant unsubscribe from your own alert. Use /delete to remove it.", nil
		}

		err := t.data.RemoveUserFromAlert(alert, update.User)

		if err != nil {
			return fmt.Sprintf("You are succesfully unsubscribed from the %s alert.", alert.Name), nil
		}

		return "", err
	})

}
