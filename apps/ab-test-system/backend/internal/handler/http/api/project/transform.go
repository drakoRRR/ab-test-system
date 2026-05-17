package project

import (
	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/project"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

func toAPIProject(p domain.Project) gen.Project {
	orgID := p.OrgID

	return gen.Project{
		Id:             p.ID,
		OrganizationId: &orgID,
		Name:           p.Name,
		Description:    &p.Description,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      &p.UpdatedAt,
	}
}

func ptr[T any](v T) *T { return &v }
