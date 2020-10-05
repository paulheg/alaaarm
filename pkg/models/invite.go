package models

import (
	"time"

	"github.com/dchest/uniuri"
)

const (
	inviteTokenLength = 16
)

// Invite represents an invitation for an alert
type Invite struct {
	AutoIncrement
	Model
	AlertID    uint `db:"alert_id"`
	Alert      *Alert
	Token      string    `db:"token"`
	OneTime    bool      `db:"one_time"`
	Expiration time.Time `db:"expiration"`
}

// NewInvite creates a new Invite instance
func NewInvite(alert *Alert) *Invite {
	invite := &Invite{}
	invite.CreatedAt.Scan(time.Now())
	invite.SetAlert(alert)
	invite.ChangeToken()

	return invite
}

// TableName returns the name of the Invite struct
func (i *Invite) TableName() string {
	return "INVITE"
}

// ChangeToken generates and sets new Token
func (i *Invite) ChangeToken() {
	i.Token = uniuri.NewLen(inviteTokenLength)
}

// SetAlert of invitation
func (i *Invite) SetAlert(alert *Alert) {
	i.Alert = alert
	i.AlertID = alert.ID
}
