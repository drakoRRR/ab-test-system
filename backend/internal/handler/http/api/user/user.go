package user

import (
	"context"
	"errors"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/user"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"
)

func (h *UserHandler) CreateUser(
	ctx context.Context,
	request gen.CreateUserRequestObject,
) (gen.CreateUserResponseObject, error) {
	if request.Body == nil {
		return gen.CreateUser400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: ptr("request body is required")},
		}, nil
	}

	uid, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return gen.CreateUser401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	u, err := h.service.CreateOrUpdate(ctx, domain.UpsertParams{
		FirebaseUID: uid,
		Email:       string(request.Body.Email),
		Name:        request.Body.Name,
	})
	if err != nil {
		return nil, err
	}

	return gen.CreateUser201JSONResponse(toAPIUser(u)), nil
}

func (h *UserHandler) GetCurrentUser(
	ctx context.Context,
	_ gen.GetCurrentUserRequestObject,
) (gen.GetCurrentUserResponseObject, error) {
	uid, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return gen.GetCurrentUser401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	u, err := h.service.GetCurrentUser(ctx, uid)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.GetCurrentUser401JSONResponse{
				UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{
					Message: ptr("user not registered"),
				},
			}, nil
		}
		return nil, err
	}

	return gen.GetCurrentUser200JSONResponse(toAPIUser(u)), nil
}
