package data

import "errors"

// Userdata interface
type Userdata interface {
	UserExists(id int64) (bool, error)
	GetAllAuthorizedUsers() ([]int64, error)
	IsUserAuthorized(id int64) (bool, error)
	DeleteUser(id int64) error
	AddUser(id int64, username string) error
}

var (
	// ErrUserAlreadyExists is returned when a user already exists
	ErrUserAlreadyExists = errors.New("user already exists")
)
