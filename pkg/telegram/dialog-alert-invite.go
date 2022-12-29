package telegram

import (
	"database/sql"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/paulheg/alaaarm/pkg/dialog"
	"github.com/paulheg/alaaarm/pkg/models"
	"github.com/skip2/go-qrcode"
)

func (t *Telegram) newAlertInviteDialog() *dialog.Dialog {

	const GROUP_CONTEXT_KEY = "Group"

	return t.newSelectAlertDialog().
		Append(t.newSelectDialog("Group", "Private", "group_private_question",
			// on group
			func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
				ctx.Set(GROUP_CONTEXT_KEY, true)
				return dialog.Next, nil
			},
			func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {
				ctx.Set(GROUP_CONTEXT_KEY, false)
				return dialog.Next, nil
			})).
		Chain(failable(func(u Update, ctx dialog.ValueStore) (dialog.Status, error) {

			alert, ok := ctx.Value(ALERT_CONTEXT_KEY).(models.Alert)
			if !ok {
				return dialog.Reset, errContextDataMissing
			}

			groupInvite, ok := ctx.Value(GROUP_CONTEXT_KEY).(bool)
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

			inviteURL := t.invitationURL(invite, groupInvite)

			qrBytes, err := qrcode.Encode(inviteURL, qrcode.Low, 200)
			if err != nil {
				return dialog.Reset, err
			}

			text, _ := u.Dictionary.Lookup("invitation_link_generated")

			msg := tgbotapi.NewPhoto(u.ChatID, tgbotapi.FileBytes{
				Name:  "invite_qr.png",
				Bytes: qrBytes,
			})
			msg.Caption = fmt.Sprintf(text, inviteURL)

			_, err = t.bot.Send(msg)
			if err != nil {
				return dialog.Reset, err
			}

			return dialog.Success, nil
		}))
}

func (t *Telegram) invitationURL(invite models.Invite, group bool) string {
	param := "start"
	if group {
		param = "startgroup"
	}

	return fmt.Sprintf("https://t.me/%s?%s=%s", t.bot.Self.UserName, param, invite.Token)
}
