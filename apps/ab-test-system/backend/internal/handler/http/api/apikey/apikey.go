package apikey

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/apikey"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"
)

func (h *APIKeyHandler) CreateApiKey(
	ctx context.Context,
	request gen.CreateApiKeyRequestObject,
) (gen.CreateApiKeyResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.CreateApiKey401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	if request.Body == nil {
		return gen.CreateApiKey400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: ptr("request body is required")},
		}, nil
	}

	key, raw, err := h.service.Create(ctx, uuid.UUID(request.ProjectId), request.Body.Name)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.CreateApiKey404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("project not found")},
			}, nil
		}

		return nil, err
	}

	return gen.CreateApiKey201JSONResponse(toAPIKeyCreated(key, raw)), nil
}

func (h *APIKeyHandler) ListApiKeys(
	ctx context.Context,
	request gen.ListApiKeysRequestObject,
) (gen.ListApiKeysResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.ListApiKeys401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	keys, err := h.service.List(ctx, uuid.UUID(request.ProjectId))
	if err != nil {
		return nil, err
	}

	resp := make(gen.ListApiKeys200JSONResponse, len(keys))
	for i, k := range keys {
		resp[i] = toAPIKey(k)
	}

	return resp, nil
}

func (h *APIKeyHandler) RevokeApiKey(
	ctx context.Context,
	request gen.RevokeApiKeyRequestObject,
) (gen.RevokeApiKeyResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.RevokeApiKey401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	err := h.service.Revoke(ctx, uuid.UUID(request.ProjectId), uuid.UUID(request.KeyId))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.RevokeApiKey404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("api key not found")},
			}, nil
		}

		return nil, err
	}

	return gen.RevokeApiKey204Response{}, nil
}
