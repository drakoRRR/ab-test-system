package experiment

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound          = errors.New("experiment not found")
	ErrConflict          = errors.New("experiment already exists")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrNotDraft          = errors.New("experiment must be in draft status")
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusRunning   Status = "running"
	StatusPaused    Status = "paused"
	StatusCompleted Status = "completed"
)

type Experiment struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	FlagID         *uuid.UUID
	Key            string
	Name           string
	Description    string
	Status         Status
	TrafficPercent float64
	Variants       []Variant
	CreatedAt      time.Time
	UpdatedAt      time.Time
	StartedAt      *time.Time
	EndedAt        *time.Time
}

type Variant struct {
	ID     uuid.UUID
	Key    string
	Name   string
	Weight int
}

var validTransitions = map[Status]map[Status]bool{
	StatusDraft:     {StatusRunning: true},
	StatusRunning:   {StatusPaused: true, StatusCompleted: true},
	StatusPaused:    {StatusRunning: true, StatusCompleted: true},
	StatusCompleted: {},
}

func (e *Experiment) CanTransitionTo(next Status) bool {
	targets, ok := validTransitions[e.Status]
	if !ok {
		return false
	}

	return targets[next]
}

type CreateParams struct {
	ProjectID      uuid.UUID
	FlagID         *uuid.UUID
	Key            string
	Name           string
	Description    string
	TrafficPercent float64
	Variants       []Variant
}

type UpdateParams struct {
	ProjectID      uuid.UUID
	ExperimentID   uuid.UUID
	Name           *string
	Description    *string
	TrafficPercent *float64
}
