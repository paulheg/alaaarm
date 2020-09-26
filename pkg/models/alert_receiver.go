package models

// AlertReceiver represents the table containing notified users
type AlertReceiver struct {
	AlertID uint `gorm:"column:alert_id"`
	UserID  uint `gorm:"column:user_id"`
}

// TableName returns the name of the AlertReceiver struct
func (ar *AlertReceiver) TableName() string {
	return "ALERT_RECEIVER"
}
