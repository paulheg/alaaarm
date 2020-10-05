package repository

import (
	"errors"
)

// Repository interface
type Repository interface {
	InitDatabase(config *Configuration) error
	MigrateDatabase() error

	AlertRepository
	InviteRepository
	UserRepository
}

var (
	// ErrUserNotFound is returned when a user was not found
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when a user already exists
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrAlreadyExists is returned when an object already exists in the database
	ErrAlreadyExists = errors.New("object already exists")
	// ErrMandatoryDataMissing is returned when data is missing to execute a certain action
	ErrMandatoryDataMissing = errors.New("mandatory data is missing")
)
