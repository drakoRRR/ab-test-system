package organization

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrAlreadyHasOrg = errors.New("user already belongs to an organization")

type Organization struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}
