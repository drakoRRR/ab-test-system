package flag

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("flag not found")
	ErrConflict = errors.New("flag key already exists")
)

type Flag struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Key       string
	Name      string
	Enabled   bool
	Rules     []Rule
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Rule struct {
	Type  string
	Value float64
}

type UpdateParams struct {
	ProjectID uuid.UUID
	Key       string
	Name      *string
	Enabled   *bool
	Rules     *[]Rule
}
