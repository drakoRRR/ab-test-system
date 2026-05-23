package flag

import (
	"context"
	"errors"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/flag"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"
)

func (h *FlagHandler) CreateFlag(
	ctx context.Context,
	request gen.CreateFlagRequestObject,
) (gen.CreateFlagResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.CreateFlag401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	if request.Body == nil {
		return gen.CreateFlag400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: ptr("request body is required")},
		}, nil
	}

	f, err := h.service.Create(ctx, request.ProjectId, request.Body.Key, request.Body.Name)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return gen.CreateFlag409JSONResponse{
				ConflictJSONResponse: gen.ConflictJSONResponse{Message: ptr("flag key already exists")},
			}, nil
		}

		return nil, err
	}

	return gen.CreateFlag201JSONResponse(toAPIFlag(f)), nil
}

func (h *FlagHandler) ListFlags(
	ctx context.Context,
	request gen.ListFlagsRequestObject,
) (gen.ListFlagsResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.ListFlags401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	flags, err := h.service.List(ctx, request.ProjectId)
	if err != nil {
		return nil, err
	}

	resp := make(gen.ListFlags200JSONResponse, len(flags))
	for i, f := range flags {
		resp[i] = toAPIFlag(f)
	}

	return resp, nil
}

func (h *FlagHandler) GetFlag(
	ctx context.Context,
	request gen.GetFlagRequestObject,
) (gen.GetFlagResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.GetFlag401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	f, err := h.service.GetByKey(ctx, request.ProjectId, request.FlagKey)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.GetFlag404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("flag not found")},
			}, nil
		}

		return nil, err
	}

	return gen.GetFlag200JSONResponse(toAPIFlag(f)), nil
}

func (h *FlagHandler) UpdateFlag(
	ctx context.Context,
	request gen.UpdateFlagRequestObject,
) (gen.UpdateFlagResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.UpdateFlag401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	if request.Body == nil {
		return gen.UpdateFlag400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: ptr("request body is required")},
		}, nil
	}

	var rules *[]domain.Rule
	if request.Body.Rules != nil {
		r := toDomainRules(*request.Body.Rules)
		rules = &r
	}

	f, err := h.service.Update(ctx, domain.UpdateParams{
		ProjectID: request.ProjectId,
		Key:       request.FlagKey,
		Name:      request.Body.Name,
		Enabled:   request.Body.Enabled,
		Rules:     rules,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.UpdateFlag404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("flag not found")},
			}, nil
		}

		return nil, err
	}

	return gen.UpdateFlag200JSONResponse(toAPIFlag(f)), nil
}

func (h *FlagHandler) DeleteFlag(
	ctx context.Context,
	request gen.DeleteFlagRequestObject,
) (gen.DeleteFlagResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.DeleteFlag401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	if err := h.service.Delete(ctx, request.ProjectId, request.FlagKey); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.DeleteFlag404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("flag not found")},
			}, nil
		}

		return nil, err
	}

	return gen.DeleteFlag204Response{}, nil
}
