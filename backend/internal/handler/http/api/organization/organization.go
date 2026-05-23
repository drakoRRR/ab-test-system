package organization

import (
	"context"
	"errors"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/organization"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"
)

func (h *Handler) CreateOrganization(
	ctx context.Context,
	request gen.CreateOrganizationRequestObject,
) (gen.CreateOrganizationResponseObject, error) {
	if request.Body == nil {
		return gen.CreateOrganization400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: ptr("request body is required")},
		}, nil
	}

	firebaseUID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return gen.CreateOrganization401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	u, err := h.userService.GetCurrentUser(ctx, firebaseUID)
	if err != nil {
		return gen.CreateOrganization401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("user not found")},
		}, nil
	}

	org, err := h.service.Create(ctx, request.Body.Name, u.ID)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyHasOrg) {
			return gen.CreateOrganization409JSONResponse{
				ConflictJSONResponse: gen.ConflictJSONResponse{Message: ptr("user already belongs to an organization")},
			}, nil
		}
		return nil, err
	}

	return gen.CreateOrganization201JSONResponse(toAPIOrg(org)), nil
}

func ptr(s string) *string { return &s }
