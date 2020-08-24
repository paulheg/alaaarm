package models

import (
	"github.com/jinzhu/gorm"
)

// User represents a Telegram user
type User struct {
	gorm.Model
	Username   string `gorm:"column:username"`
	TelegramID int64  `gorm:"colum:telegram_id;unique"`
}

// TableName returns the table name of the User struct
func (u *User) TableName() string {
	return "USER"
}

// NewUser creates a new user struct
func NewUser(telegramID int64, username string) User {
	return User{
		Username:   username,
		TelegramID: telegramID,
	}
}
