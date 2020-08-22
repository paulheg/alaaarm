package telegram

import (
	"fmt"

	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
)

func (t *Telegram) newAlertInviteDialog() *dialog.Dialog {

	return t.newSelectDialog("invite", func(alert models.Alert) (string, error) {
		invite, err := t.data.CreateInvitation(alert)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Here is your invitation link:\n\n%s", t.invitationURL(invite)), nil
	})
}

func (t *Telegram) invitationURL(invite models.Invite) string {
	return fmt.Sprintf("https://t.me/%s?start=%s", t.config.Name, invite.Token)
}
