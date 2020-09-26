package models

import (
	"time"

	"github.com/dchest/uniuri"
	"gorm.io/gorm"
)

const (
	inviteTokenLength = 16
)

// Invite represents an invitation for an alert
type Invite struct {
	gorm.Model
	AlertID    uint
	Alert      Alert
	Token      string
	OneTime    bool
	Expiration time.Time
}

// TableName returns the name of the Invite struct
func (i *Invite) TableName() string {
	return "INVITE"
}

// GenerateToken generates a new token
func (i *Invite) GenerateToken() {
	i.Token = uniuri.NewLen(inviteTokenLength)
}
