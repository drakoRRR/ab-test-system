package event

import (
	"time"

	"github.com/google/uuid"
)

type Type string

const (
	TypeExposure   Type = "exposure"
	TypeConversion Type = "conversion"
)

type Event struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	UserID       string
	ExperimentID uuid.UUID
	VariantID    uuid.UUID
	Type         Type
	Name         string
	Value        float64
	Timestamp    time.Time
}
