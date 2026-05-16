package apikey

import (
	openapi_types "github.com/oapi-codegen/runtime/types"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/apikey"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

func toAPIKey(k domain.Key) gen.ApiKey {
	return gen.ApiKey{
		Id:        openapi_types.UUID(k.ID),
		ProjectId: openapi_types.UUID(k.ProjectID),
		Name:      k.Name,
		Prefix:    k.Prefix,
		CreatedAt: k.CreatedAt,
		RevokedAt: k.RevokedAt,
	}
}

func toAPIKeyCreated(k domain.Key, rawKey string) gen.ApiKeyCreated {
	return gen.ApiKeyCreated{
		Id:        openapi_types.UUID(k.ID),
		ProjectId: openapi_types.UUID(k.ProjectID),
		Name:      k.Name,
		Prefix:    k.Prefix,
		Key:       rawKey,
		CreatedAt: k.CreatedAt,
		RevokedAt: k.RevokedAt,
	}
}

func ptr[T any](v T) *T { return &v }
