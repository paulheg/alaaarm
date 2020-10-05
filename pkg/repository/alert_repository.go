package repository

import "github.com/paulheg/alaaarm/pkg/models"

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
