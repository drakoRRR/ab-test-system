package project

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("project not found")

type Project struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
