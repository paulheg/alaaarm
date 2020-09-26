package models

// AlertReceiver represents the table containing notified users
type AlertReceiver struct {
	AlertID uint
	UserID  uint
}

// TableName returns the name of the AlertReceiver struct
func (ar *AlertReceiver) TableName() string {
	return "ALERT_RECEIVER"
}
