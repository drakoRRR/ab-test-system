package sdk

import (
	"context"

	sdkgen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen/sdk"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"
)

func (h *Handler) GetSdkConfig(
	ctx context.Context,
	_ sdkgen.GetSdkConfigRequestObject,
) (sdkgen.GetSdkConfigResponseObject, error) {
	projectID, ok := middleware.ProjectIDFromContext(ctx)
	if !ok {
		return sdkgen.GetSdkConfig401JSONResponse{
			UnauthorizedJSONResponse: sdkgen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	cfg, err := h.service.GetConfig(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return sdkgen.GetSdkConfig200JSONResponse(toAPIConfig(cfg)), nil
}
