package telegram

import (
	"database/sql"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
	"github.com/skip2/go-qrcode"
)

func (t *Telegram) newAlertInviteDialog() *dialog.Dialog {

	return t.newSelectAlertDialog().
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value("alert").(models.Alert)

			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			invite, err := t.repository.GetInviteByAlertID(alert.ID)
			if err == sql.ErrNoRows {
				invite, err = t.repository.CreateInvite(alert)
				if err != nil {
					return dialog.Reset, err
				}
			} else if err != nil {
				return dialog.Reset, err
			}

			inviteURL := t.invitationURL(invite)

			qrBytes, err := qrcode.Encode(inviteURL, qrcode.Low, 200)
			if err != nil {
				return dialog.Reset, err
			}

			msg := tgbotapi.NewPhotoUpload(u.ChatID, tgbotapi.FileBytes{
				Name:  "invite_qr.png",
				Bytes: qrBytes,
			})
			msg.Caption = fmt.Sprintf("Here is your invitation link:\n\n%s", inviteURL)

			_, err = t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		}))
}

func (t *Telegram) invitationURL(invite models.Invite) string {
	return fmt.Sprintf("https://t.me/%s?start=%s", t.bot.Self.UserName, invite.Token)
}
