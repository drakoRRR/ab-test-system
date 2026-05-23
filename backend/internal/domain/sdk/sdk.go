package sdk

import (
	"github.com/google/uuid"

	domainexp "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
	domainflag "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/flag"
)

type Config struct {
	ProjectID   uuid.UUID
	Flags       []domainflag.Flag
	Experiments []domainexp.Experiment
}
