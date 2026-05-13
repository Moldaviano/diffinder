package model

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleDeveloper UserRole = "developer"
	RoleViewer    UserRole = "viewer"
)

func (r UserRole) Valid() bool {
	switch r {
	case RoleAdmin, RoleDeveloper, RoleViewer:
		return true
	}
	return false
}

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
