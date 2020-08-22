package models

import (
	"github.com/jinzhu/gorm"
)

const (
	tokenLength = 32
)

// Alert represents a custom alert to send notifications to
type Alert struct {
	gorm.Model
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	OwnerID     uint   `gorm:"column:owner_id"`
	Owner       User   `gorm:"foreignkey:owner_id"`
	Token       string `gorm:"column:token"`
	Receiver    []User `gorm:"many2many:ALERT_RECEIVER"`
}

// TableName returns the name of the Hook struct
func (a *Alert) TableName() string {
	return "ALERT"
}
