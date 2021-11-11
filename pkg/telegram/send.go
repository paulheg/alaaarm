package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/paulheg/alaaarm/pkg/models"
)

type chattableCreate func(id int64, file tgbotapi.FileReader, message string) tgbotapi.Chattable
type chattableShare func(id int64, fileID string, message string) tgbotapi.Chattable
type getFileKey func(msg tgbotapi.Message) (string, error)

func (t *Telegram) shareToAll(receivers []models.User, file tgbotapi.FileReader, message string, create chattableCreate, fileID getFileKey, share chattableShare) error {

	first := receivers[0]
	createFile := create(first.TelegramID, file, message)

	msg, err := t.bot.Send(createFile)
	if err != nil {
		return err
	}

	key, err := fileID(msg)
	if err != nil {
		return err
	}

	for _, receiver := range receivers[1:] {
		chattable := share(receiver.TelegramID, key, message)

		_, err = t.bot.Send(chattable)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Telegram) sendImageToAll(alert models.Alert, image tgbotapi.FileReader, message string) error {

	receivers, err := t.repository.GetAlertReceiver(alert)
	if err != nil {
		return err
	}

	receivers = append(receivers, alert.Owner)

	create := func(id int64, file tgbotapi.FileReader, message string) tgbotapi.Chattable {
		foto := tgbotapi.NewPhoto(id, file)
		foto.Caption = message

		return foto
	}

	fileID := func(msg tgbotapi.Message) (string, error) {

		if msg.Photo != nil {
			photos := msg.Photo

			if len(photos) > 0 {
				return photos[0].FileID, nil
			}
		}

		return "", errPhotoIDMissing
	}

	share := func(id int64, fileID string, messsage string) tgbotapi.Chattable {
		foto := tgbotapi.NewPhoto(id, tgbotapi.FileID(fileID))
		foto.Caption = message

		return foto
	}

	return t.shareToAll(receivers, image, message, create, fileID, share)
}

func (t *Telegram) sendDocumentToAll(alert models.Alert, document tgbotapi.FileReader, message string) error {

	receivers, err := t.repository.GetAlertReceiver(alert)
	if err != nil {
		return err
	}

	receivers = append(receivers, alert.Owner)

	create := func(id int64, file tgbotapi.FileReader, message string) tgbotapi.Chattable {
		document := tgbotapi.NewDocument(id, file)
		document.Caption = message

		return document
	}

	fileID := func(msg tgbotapi.Message) (string, error) {
		if msg.Document != nil {
			return msg.Document.FileID, nil
		}
		return "", errDocumentIDMissing
	}

	share := func(id int64, fileID string, message string) tgbotapi.Chattable {
		document := tgbotapi.NewDocument(id, tgbotapi.FileID(fileID))
		document.Caption = message

		return document
	}

	return t.shareToAll(receivers, document, message, create, fileID, share)
}

func (t *Telegram) sendMessageToAll(alert models.Alert, message string) error {
	receivers, err := t.repository.GetAlertReceiver(alert)
	if err != nil {
		return err
	}

	receivers = append(receivers, alert.Owner)

	for _, receiver := range receivers {
		msg := tgbotapi.NewMessage(receiver.TelegramID, message)
		_, err := t.bot.Send(msg)

		if err != nil {
			if terr, ok := err.(*tgbotapi.Error); ok {
				switch terr.Code {
				case 403:
					t.log.WithField("userID", receiver.ID).Warn("Not allowed to send message. Deleting user...")
					err = t.repository.DeleteUser(receiver.ID)
					if err != nil {
						return err
					}
				default:
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}
