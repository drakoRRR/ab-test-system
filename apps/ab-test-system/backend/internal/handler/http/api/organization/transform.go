package organization

import (
	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/organization"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

func toAPIOrg(o domain.Organization) gen.Organization {
	return gen.Organization{
		Id:        o.ID,
		Name:      o.Name,
		CreatedAt: o.CreatedAt,
	}
}
