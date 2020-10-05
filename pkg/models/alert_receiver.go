package models

import (
	"time"
)

// AlertReceiver represents the table containing notified users
type AlertReceiver struct {
	AlertID uint `db:"alert_id"`
	UserID  uint `db:"user_id"`
	Model
}

// NewAlertReceiver creates a new AlertReceiver instance
func NewAlertReceiver(user *User, alert *Alert) *AlertReceiver {
	alertReceiver := &AlertReceiver{
		AlertID: alert.ID,
		UserID:  user.ID,
	}
	alertReceiver.CreatedAt.Scan(time.Now())

	return alertReceiver
}

// TableName returns the name of the AlertReceiver struct
func (ar *AlertReceiver) TableName() string {
	return "ALERT_RECEIVER"
}
