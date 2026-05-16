package apikey

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("api key not found")

type Key struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Name      string
	KeyHash   string
	Prefix    string
	CreatedAt time.Time
	RevokedAt *time.Time
}
