package analytics

import (
	"context"
	"errors"

	domainexperiment "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

func (h *AnalyticsHandler) GetExperimentAnalytics(
	ctx context.Context,
	request gen.GetExperimentAnalyticsRequestObject,
) (gen.GetExperimentAnalyticsResponseObject, error) {
	result, err := h.service.GetResult(ctx, request.ProjectId, request.ExperimentId)
	if err != nil {
		if errors.Is(err, domainexperiment.ErrNotFound) {
			return gen.GetExperimentAnalytics404JSONResponse{}, nil
		}
		return nil, err
	}

	return gen.GetExperimentAnalytics200JSONResponse(toAPIAnalytics(result)), nil
}
