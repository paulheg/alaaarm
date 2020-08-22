package models

import "github.com/jinzhu/gorm"

// AlertReceiver represents the table containing notified users
type AlertReceiver struct {
	gorm.Model
	AlertID uint
	UserID  uint
}

// TableName returns the name of the AlertReceiver struct
func (ar *AlertReceiver) TableName() string {
	return "ALERT_RECEIVER"
}
