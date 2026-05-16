package project

import (
	openapi_types "github.com/oapi-codegen/runtime/types"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/project"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

func toAPIProject(p domain.Project) gen.Project {
	orgID := openapi_types.UUID(p.OrgID)

	return gen.Project{
		Id:             openapi_types.UUID(p.ID),
		OrganizationId: &orgID,
		Name:           p.Name,
		Description:    &p.Description,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      &p.UpdatedAt,
	}
}

func ptr[T any](v T) *T { return &v }
