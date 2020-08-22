package data

import (
	"errors"

	"github.com/paulheg/alaaarm/pkg/models"
)

// Data interface
type Data interface {
	InitDatabase() error
	MigrateDatabase() error

	// Invitation Related
	CreateInvitation(alert models.Alert) (models.Invite, error)
	GetInvitation(token string) (bool, models.Invite, error)

	// Alert related
	CreateAlert(name, description string, owner models.User) (models.Alert, error)
	DeleteAlert(alert models.Alert) error
	GetAlertWithToken(token string) (bool, models.Alert, error)
	GetAlert(id uint) (bool, models.Alert, error)
	UpdateAlertToken(alert models.Alert) (models.Alert, error)

	GetUserAlerts(userID uint) ([]models.Alert, error)
	GetUserTelegramAlerts(telegramID int64) ([]models.Alert, error)
	AddUserToAlert(alert models.Alert, user models.User) error
	RemoveUserFromAlert(alert models.Alert, user models.User) error

	// User functions
	CreateUser(user models.User) (models.User, error)
	GetUser(userID uint) (bool, models.User, error)
	GetUserTelegram(telegramID int64) (bool, models.User, error)
	DeleteUser(userID uint) error
}

var (
	// ErrUserNotFound is returned when a user was not found
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when a user already exists
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrAlreadyExists is returned when an object already exists in the database
	ErrAlreadyExists = errors.New("object already exists")
	// ErrMandatoryDataMissing is returned when data is missing to execute a certain action
	ErrMandatoryDataMissing = errors.New("mandatory data is missing")
)
