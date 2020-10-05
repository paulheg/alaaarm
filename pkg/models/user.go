package models

import (
	"fmt"
	"time"
)

// User represents a Telegram user
type User struct {
	AutoIncrement
	Model
	Username   string `db:"username"`
	TelegramID int64  `db:"telegram_id"`
}

// TableName returns the table name of the User struct
func (u *User) TableName() string {
	return "USER"
}

// TelegramUserLink to reference users
func (u *User) TelegramUserLink() string {
	return fmt.Sprintf("tg://user?id=%v", u.TelegramID)
}

// NewUser creates a new User instance
func NewUser(telegramID int64, username string) *User {
	user := &User{
		Username:   username,
		TelegramID: telegramID,
	}

	user.CreatedAt.Scan(time.Now())

	return user
}
