package models

import (
	"time"

	"github.com/jinzhu/gorm"
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
