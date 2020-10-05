package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/paulheg/alaaarm/pkg/models"
)

type chattableCreate func(id int64, file tgbotapi.FileReader) tgbotapi.Chattable
type chattableShare func(id int64, fileID string) tgbotapi.Chattable
type getFileKey func(msg tgbotapi.Message) (string, error)

func (t *Telegram) shareToAll(receivers []models.User, file tgbotapi.FileReader, create chattableCreate, fileID getFileKey, share chattableShare) error {

	first := receivers[0]
	createFile := create(first.TelegramID, file)

	msg, err := t.bot.Send(createFile)
	if err != nil {
		return err
	}

	key, err := fileID(msg)
	if err != nil {
		return err
	}

	for _, receiver := range receivers[1:] {
		chattable := share(receiver.TelegramID, key)

		_, err = t.bot.Send(chattable)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Telegram) sendImage(alert models.Alert, image tgbotapi.FileReader) error {

	receivers, err := t.repository.GetAlertReceiver(alert)
	if err != nil {
		return err
	}

	receivers = append(receivers, alert.Owner)

	create := func(id int64, file tgbotapi.FileReader) tgbotapi.Chattable {
		return tgbotapi.NewPhotoUpload(id, file)
	}

	fileID := func(msg tgbotapi.Message) (string, error) {

		if msg.Photo != nil {
			photos := (*msg.Photo)

			if len(photos) > 0 {
				return photos[0].FileID, nil
			}
		}

		return "", errPhotoIDMissing
	}

	share := func(id int64, fileID string) tgbotapi.Chattable {
		return tgbotapi.NewPhotoShare(id, fileID)
	}

	return t.shareToAll(receivers, image, create, fileID, share)
}

func (t *Telegram) sendDocument(alert models.Alert, document tgbotapi.FileReader) error {

	receivers, err := t.repository.GetAlertReceiver(alert)
	if err != nil {
		return err
	}

	receivers = append(receivers, alert.Owner)

	create := func(id int64, file tgbotapi.FileReader) tgbotapi.Chattable {
		return tgbotapi.NewDocumentUpload(id, file)
	}

	fileID := func(msg tgbotapi.Message) (string, error) {
		if msg.Document != nil {
			return msg.Document.FileID, nil
		}
		return "", errDocumentIDMissing
	}

	share := func(id int64, fileID string) tgbotapi.Chattable {
		return tgbotapi.NewDocumentShare(id, fileID)
	}

	return t.shareToAll(receivers, document, create, fileID, share)
}

func (t *Telegram) sendToAll(alert models.Alert, message string) error {
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
					log.Printf("Not allowed to send user %x messages. Deleting user...", receiver.ID)
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
