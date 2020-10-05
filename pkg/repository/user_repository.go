package repository

import "github.com/paulheg/alaaarm/pkg/models"

type UserRepository interface {
	CreateUser(user models.User) (models.User, error)
	GetUser(userID uint) (models.User, error)
	GetUserTelegram(telegramID int64) (models.User, error)
	DeleteUser(userID uint) error
}
