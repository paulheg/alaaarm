package repository

import "github.com/paulheg/alaaarm/pkg/models"

// InviteRepository defines invite related data operations
type InviteRepository interface {
	CreateInvite(alert models.Alert) (models.Invite, error)
	GetInviteByToken(token string) (models.Invite, error)
	GetInvite(inviteID uint) (models.Invite, error)
	GetInviteByAlertID(alertID uint) (models.Invite, error)
	DeleteInvite(inviteID uint) error
}
