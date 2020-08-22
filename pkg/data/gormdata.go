package data

import (
	"github.com/dchest/uniuri"
	"github.com/jinzhu/gorm"

	// sqlite driver
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/paulheg/alaaarm/pkg/models"
)

type gormdata struct {
	db     *gorm.DB
	config Configuration
}

const (
	inviteTokenLength = 16
	alertTokenLength  = 32
)

// NewGormData return Data interface with gorm
func NewGormData(config Configuration) Data {
	return &gormdata{
		config: config,
	}

}

func (d *gormdata) MigrateDatabase() error {
	var err error

	tables := make([]interface{}, 0)

	tables = append(tables, models.User{})
	tables = append(tables, models.Alert{})
	tables = append(tables, models.AlertReceiver{})
	tables = append(tables, models.Invite{})

	for _, table := range tables {
		if !d.db.HasTable(table) {
			d.db.CreateTable(table)
		}
	}

	result := d.db.AutoMigrate(tables...)
	err = result.Error

	return err
}

func (d *gormdata) InitDatabase() error {
	var err error

	// database setup
	db, err := gorm.Open("sqlite3", d.config.ConnectionString)
	d.db = db

	return err
}

// === ALERTS ===

func (d *gormdata) CreateAlert(name, description string, owner models.User) (models.Alert, error) {

	// check inputs
	if len(name) <= 0 {
		return models.Alert{}, ErrMandatoryDataMissing
	}

	if len(description) <= 0 {
		return models.Alert{}, ErrMandatoryDataMissing
	}

	token := uniuri.NewLen(alertTokenLength)

	alert := models.Alert{
		Name:        name,
		Description: description,
		Owner:       owner,
		OwnerID:     owner.ID,
		Receiver:    make([]models.User, 0),
		Token:       token,
	}

	if d.db.NewRecord(alert) {
		result := d.db.Create(&alert)
		return alert, result.Error
	}
	return alert, ErrAlreadyExists
}

func (d *gormdata) DeleteAlert(alert models.Alert) error {
	result := d.db.Delete(&alert)

	return result.Error
}

func (d *gormdata) GetAlertWithToken(token string) (bool, models.Alert, error) {
	var alert models.Alert
	//var user models.User

	result := d.db.First(&alert, "Token = ?", token).
		Related(&alert.Owner, "OwnerID").
		Related(&alert.Receiver, "Receiver")

	return !result.RecordNotFound(), alert, result.Error
}

func (d *gormdata) GetAlert(id uint) (bool, models.Alert, error) {
	var alert models.Alert

	result := d.db.First(&alert, "ID = ?", id).
		Related(&alert.Owner, "OwnerID").
		Related(&alert.Receiver, "Receiver")

	return !result.RecordNotFound(), alert, result.Error
}

func (d *gormdata) RemoveUserFromAlert(alert models.Alert, user models.User) error {
	result := d.db.Model(&alert).Association("Receiver").Delete(&user)
	return result.Error
}

func (d *gormdata) AddUserToAlert(alert models.Alert, user models.User) error {
	result := d.db.Model(&alert).Association("Receiver").Append(&user)
	return result.Error
}

// === USER ====

func (d *gormdata) CreateUser(user models.User) (models.User, error) {
	result := d.db.Create(&user)
	return user, result.Error
}

func (d *gormdata) GetUser(userID uint) (bool, models.User, error) {
	var user models.User

	result := d.db.Find(&user, "ID = ?", userID)

	return !result.RecordNotFound(), user, result.Error
}

func (d *gormdata) GetUserTelegram(telegramID int64) (bool, models.User, error) {
	var user models.User

	result := d.db.Find(&user, "Telegram_ID = ?", telegramID)

	return !result.RecordNotFound(), user, result.Error
}

func (d *gormdata) DeleteUser(userID uint) error {
	user := models.User{
		Model: gorm.Model{ID: uint(userID)},
	}
	result := d.db.Delete(&user)
	return result.Error
}

func (d *gormdata) GetUserAlerts(userID uint) ([]models.Alert, error) {
	var alerts []models.Alert

	result := d.db.Find(&alerts, "owner_id = ?", userID)

	return alerts, result.Error
}

func (d *gormdata) GetUserTelegramAlerts(telegramID int64) ([]models.Alert, error) {
	var alerts []models.Alert

	var query string = `SELECT *
FROM ALERT
INNER JOIN USER ON ALERT.owner_id = USER.id
WHERE USER.telegram_id = ?
AND ALERT.deleted_at IS NULL`

	result := d.db.Raw(query, telegramID).Scan(&alerts)

	return alerts, result.Error
}

func (d *gormdata) UpdateAlertToken(alert models.Alert) (models.Alert, error) {
	newToken := uniuri.NewLen(alertTokenLength)

	alert.Token = newToken
	result := d.db.Update(alert)

	return alert, result.Error
}

// === INVITES ===

func (d *gormdata) CreateInvitation(alert models.Alert) (models.Invite, error) {
	token := uniuri.NewLen(inviteTokenLength)

	invite := models.Invite{
		Alert: alert,
		Token: token,
	}

	result := d.db.Create(&invite)

	return invite, result.Error
}

func (d *gormdata) GetInvitation(token string) (bool, models.Invite, error) {
	var invite models.Invite

	result := d.db.First(&invite, "Token = ?", token).
		Related(&invite.Alert, "AlertID").
		Related(&invite.Alert.Owner, "OwnerID")

	return !result.RecordNotFound(), invite, result.Error
}
