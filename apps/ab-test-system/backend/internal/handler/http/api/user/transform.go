package user

import (
	openapi_types "github.com/oapi-codegen/runtime/types"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/user"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

func toAPIUser(u domain.User) gen.User {
	role := gen.UserRole(u.Role)

	out := gen.User{
		Id:        u.ID.String(),
		Email:     openapi_types.Email(u.Email),
		Name:      u.Name,
		PhotoURL:  u.PhotoURL,
		Role:      &role,
		CreatedAt: &u.CreatedAt,
		UpdatedAt: &u.UpdatedAt,
	}

	if u.OrgID != nil {
		orgID := *u.OrgID
		out.OrganizationId = &orgID
	}

	return out
}

func ptr[T any](v T) *T { return &v }
