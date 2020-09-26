package data

import (
	"gorm.io/driver/postgres" // sqlite driver
	"gorm.io/gorm"

	"github.com/paulheg/alaaarm/pkg/models"
)

type gormdata struct {
	db     *gorm.DB
	config *Configuration
}

// NewGormData return Data interface with gorm
func NewGormData(config *Configuration) Data {
	return &gormdata{
		config: config,
	}

}

func (d *gormdata) MigrateDatabase() error {
	tables := make([]interface{}, 0)

	tables = append(tables, models.User{})
	tables = append(tables, models.Alert{})
	tables = append(tables, models.AlertReceiver{})
	tables = append(tables, models.Invite{})

	return d.db.AutoMigrate(tables...)
}

func (d *gormdata) InitDatabase() error {
	var err error

	// database setup
	db, err := gorm.Open(postgres.Open(d.config.ConnectionString), &gorm.Config{})
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

	alert := models.Alert{
		Name:        name,
		Description: description,
		Owner:       owner,
		OwnerID:     owner.ID,
		Receiver:    make([]models.User, 0),
	}

	alert.ChangeToken()

	result := d.db.Create(&alert)

	return alert, result.Error
}

func (d *gormdata) DeleteAlert(alert models.Alert) error {
	result := d.db.Delete(&alert)

	if result.Error != nil {
		return result.Error
	}

	// Clear invitations
	var invite models.Invite
	result = d.db.Delete(&invite, "id = ?", alert.ID)

	if result.Error != nil {
		return result.Error
	}

	// Clear receiver
	var alertReceiver models.AlertReceiver
	result = d.db.Delete(&alertReceiver).Where("user_id", alert.OwnerID)

	return result.Error
}

func (d *gormdata) GetAlertWithToken(token string) (bool, models.Alert, error) {
	var alert models.Alert
	//var user models.User

	result := d.db.First(&alert, "Token = ?", token).Preload("Receiver")

	return result.RowsAffected > 0, alert, result.Error
}

func (d *gormdata) GetAlert(id uint) (bool, models.Alert, error) {
	var alert models.Alert

	result := d.db.First(&alert).Preload("Receiver").Preload("Owner")

	return result.RowsAffected > 0, alert, result.Error
}

func (d *gormdata) RemoveUserFromAlert(alert models.Alert, user models.User) error {
	return d.db.Model(&alert).Association("Receiver").Delete(&user)
}

func (d *gormdata) AddUserToAlert(alert models.Alert, user models.User) error {
	return d.db.Model(&alert).Association("Receiver").Append(&user)
}

// === USER ====

func (d *gormdata) CreateUser(user models.User) (models.User, error) {
	result := d.db.Create(&user)
	return user, result.Error
}

func (d *gormdata) GetUser(userID uint) (bool, models.User, error) {
	var user models.User

	result := d.db.Find(&user, "id = ?", userID)

	return result.RowsAffected > 0, user, result.Error
}

func (d *gormdata) GetUserTelegram(telegramID int64) (bool, models.User, error) {
	var user models.User

	result := d.db.Find(&user, "Telegram_ID = ?", telegramID)

	return result.RowsAffected > 0, user, result.Error
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

	result := d.db.Find(&alerts, "owner_id = ?", userID).Preload("USER")

	return alerts, result.Error
}

func (d *gormdata) GetUserSubscribedAlerts(userID uint) ([]models.Alert, error) {
	var alerts []models.Alert

	result := d.db.Find(&alerts).
		Joins("INNER JOIN ALERT_RECEIVER on ALERT.id = ALERT_RECEIVER.alert_id").
		Where("ALERT_RECEIVER.user_id = ?", userID)

	return alerts, result.Error
}

func (d *gormdata) GetUserTelegramAlerts(telegramID int64) ([]models.Alert, error) {
	var alerts []models.Alert

	result := d.db.Find(&alerts).
		Joins("INNER JOIN USER on USER.id = ALERT.owner_id").
		Where("USER.telegram_id", telegramID)

	return alerts, result.Error
}

func (d *gormdata) UpdateAlertToken(alert models.Alert) (models.Alert, error) {

	alert.ChangeToken()

	result := d.db.Update("token", &alert)

	return alert, result.Error
}

// === INVITES ===

func (d *gormdata) CreateInvitation(alert models.Alert) (models.Invite, error) {

	invite := models.Invite{
		Alert: alert,
	}

	invite.GenerateToken()

	result := d.db.Create(&invite)

	return invite, result.Error
}

func (d *gormdata) GetInvitation(token string) (bool, models.Invite, error) {
	var invite models.Invite

	result := d.db.First(&invite).
		Where("ALERT.token = ?", token).
		Preload("Alert")

	return result.RowsAffected > 0, invite, result.Error
}
