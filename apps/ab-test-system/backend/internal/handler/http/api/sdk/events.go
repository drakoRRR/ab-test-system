package sdk

import (
	"context"

	sdkgen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen/sdk"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"
)

func (h *Handler) PostSdkEvents(
	ctx context.Context,
	request sdkgen.PostSdkEventsRequestObject,
) (sdkgen.PostSdkEventsResponseObject, error) {
	projectID, ok := middleware.ProjectIDFromContext(ctx)
	if !ok {
		return sdkgen.PostSdkEvents401JSONResponse{
			UnauthorizedJSONResponse: sdkgen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	if request.Body == nil {
		return sdkgen.PostSdkEvents400JSONResponse{
			BadRequestJSONResponse: sdkgen.BadRequestJSONResponse{Message: ptr("request body is required")},
		}, nil
	}

	events := toDomainEvents(projectID, request.Body.Events)
	if err := h.events.Ingest(ctx, events); err != nil {
		return nil, err
	}

	return sdkgen.PostSdkEvents202Response{}, nil
}
