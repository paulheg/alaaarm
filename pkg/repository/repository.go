package repository

import (
	"errors"

	"github.com/paulheg/alaaarm/pkg/migration"
	"github.com/paulheg/alaaarm/pkg/models"
)

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

// Repository interface
type Repository interface {
	InitDatabase(config *Configuration) error
	MigrateDatabase() error

	AlertRepository
	InviteRepository
	UserRepository

	migration.VersionRepository
}

// UserRepository defines userdata operations
type UserRepository interface {
	CreateUser(user models.User) (models.User, error)
	GetUser(userID uint) (models.User, error)
	GetUserTelegram(telegramID int64) (models.User, error)
	DeleteUser(userID uint) error
}

// AlertRepository defines Alert related data operations
type AlertRepository interface {
	CreateAlert(alert models.Alert) (models.Alert, error)
	DeleteAlert(alert models.Alert) error
	GetAlertByToken(token string) (models.Alert, error)
	GetAlert(id uint) (models.Alert, error)
	UpdateAlertToken(alert models.Alert) (models.Alert, error)
	GetAlertReceiver(alert models.Alert) ([]models.User, error)

	GetUserAlerts(userID uint) ([]models.Alert, error)
	GetUserSubscribedAlerts(userID uint) ([]models.Alert, error)

	AddUserToAlert(alert models.Alert, user models.User) error
	RemoveUserFromAlert(alert models.Alert, user models.User) error
}

// InviteRepository defines invite related data operations
type InviteRepository interface {
	CreateInvite(alert models.Alert) (models.Invite, error)
	GetInviteByToken(token string) (models.Invite, error)
	GetInvite(inviteID uint) (models.Invite, error)
	GetInviteByAlertID(alertID uint) (models.Invite, error)
	DeleteInvite(inviteID uint) error
}
