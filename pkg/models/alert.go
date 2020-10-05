package models

import (
	"time"

	"github.com/dchest/uniuri"
)

const (
	alertTokenLength = 32
)

// Alert represents a custom alert to send notifications to
type Alert struct {
	AutoIncrement
	Model
	Name        string `db:"name"`
	Description string `db:"description"`
	OwnerID     uint   `db:"owner_id"`
	Owner       User
	Token       string `db:"token"`
}

// NewAlert creates a new Alert instance
func NewAlert(name, description string, owner User) *Alert {
	alert := &Alert{
		Name:        name,
		Description: description,
	}
	alert.CreatedAt.Scan(time.Now())
	alert.SetOwner(owner)
	alert.ChangeToken()

	return alert
}

// TableName returns the name of the Hook struct
func (a *Alert) TableName() string {
	return "ALERT"
}

// ChangeToken generates and sets new Token
func (a *Alert) ChangeToken() {
	newToken := uniuri.NewLen(alertTokenLength)
	a.Token = newToken
}

// SetOwner updates the Owner and OwnerID field
func (a *Alert) SetOwner(owner User) {
	a.Owner = owner
	a.OwnerID = owner.ID
}
