package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("user not found")
	ErrConflict = errors.New("user already exists")
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleViewer Role = "viewer"
)

type User struct {
	ID          uuid.UUID
	FirebaseUID string
	OrgID       *uuid.UUID
	Email       string
	Name        string
	PhotoURL    *string
	Role        Role
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UpsertParams struct {
	FirebaseUID string
	Email       string
	Name        string
	PhotoURL    *string
}
