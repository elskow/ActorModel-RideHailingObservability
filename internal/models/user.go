package models

import (
	"time"

	"github.com/google/uuid"
)

// UserType represents the type of user
type UserType string

const (
	UserTypePassenger UserType = "passenger"
	UserTypeDriver    UserType = "driver"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Phone     string    `json:"phone" gorm:"uniqueIndex;not null"`
	Name      string    `json:"name" gorm:"not null"`
	UserType  UserType  `json:"user_type" gorm:"not null;check:user_type IN ('passenger', 'driver')"`
	CreatedAt time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for User
func (User) TableName() string {
	return "users"
}

// IsDriver returns true if the user is a driver
func (u *User) IsDriver() bool {
	return u.UserType == UserTypeDriver
}

// IsPassenger returns true if the user is a passenger
func (u *User) IsPassenger() bool {
	return u.UserType == UserTypePassenger
}

// Validate validates the user data
func (u *User) Validate() error {
	if u.Email == "" {
		return ErrInvalidEmail
	}
	if u.Phone == "" {
		return ErrInvalidPhone
	}
	if u.Name == "" {
		return ErrInvalidName
	}
	if u.UserType != UserTypePassenger && u.UserType != UserTypeDriver {
		return ErrInvalidUserType
	}
	return nil
}
