package experiment

import (
	"context"
	"errors"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"
)

func (h *ExperimentHandler) CreateExperiment(
	ctx context.Context,
	request gen.CreateExperimentRequestObject,
) (gen.CreateExperimentResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.CreateExperiment401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	if request.Body == nil {
		return gen.CreateExperiment400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: ptr("request body is required")},
		}, nil
	}

	exp, err := h.service.Create(ctx, domain.CreateParams{
		ProjectID:      request.ProjectId,
		FlagID:         request.Body.FlagId,
		Name:           request.Body.Name,
		Description:    derefString(request.Body.Description),
		TrafficPercent: float64(request.Body.TrafficPercent),
		Variants:       toDomainVariants(request.Body.Variants),
	})
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return gen.CreateExperiment409JSONResponse{
				ConflictJSONResponse: gen.ConflictJSONResponse{Message: ptr("experiment already exists")},
			}, nil
		}

		return nil, err
	}

	return gen.CreateExperiment201JSONResponse(toAPIExperiment(exp)), nil
}

func (h *ExperimentHandler) ListExperiments(
	ctx context.Context,
	request gen.ListExperimentsRequestObject,
) (gen.ListExperimentsResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.ListExperiments401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	experiments, err := h.service.List(ctx, request.ProjectId)
	if err != nil {
		return nil, err
	}

	resp := make(gen.ListExperiments200JSONResponse, len(experiments))
	for i, exp := range experiments {
		resp[i] = toAPIExperiment(exp)
	}

	return resp, nil
}

func (h *ExperimentHandler) GetExperiment(
	ctx context.Context,
	request gen.GetExperimentRequestObject,
) (gen.GetExperimentResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.GetExperiment401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	exp, err := h.service.GetByID(ctx, request.ProjectId, request.ExperimentId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return gen.GetExperiment404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("experiment not found")},
			}, nil
		}

		return nil, err
	}

	return gen.GetExperiment200JSONResponse(toAPIExperiment(exp)), nil
}

func (h *ExperimentHandler) UpdateExperiment(
	ctx context.Context,
	request gen.UpdateExperimentRequestObject,
) (gen.UpdateExperimentResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.UpdateExperiment401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	if request.Body == nil {
		return gen.UpdateExperiment400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: ptr("request body is required")},
		}, nil
	}

	var trafficPercent *float64
	if request.Body.TrafficPercent != nil {
		v := float64(*request.Body.TrafficPercent)
		trafficPercent = &v
	}

	exp, err := h.service.Update(ctx, domain.UpdateParams{
		ProjectID:      request.ProjectId,
		ExperimentID:   request.ExperimentId,
		Name:           request.Body.Name,
		Description:    request.Body.Description,
		TrafficPercent: trafficPercent,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return gen.UpdateExperiment404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("experiment not found")},
			}, nil
		case errors.Is(err, domain.ErrNotDraft):
			return gen.UpdateExperiment409JSONResponse{
				ConflictJSONResponse: gen.ConflictJSONResponse{Message: ptr("experiment must be in draft status")},
			}, nil
		}

		return nil, err
	}

	return gen.UpdateExperiment200JSONResponse(toAPIExperiment(exp)), nil
}

func (h *ExperimentHandler) DeleteExperiment(
	ctx context.Context,
	request gen.DeleteExperimentRequestObject,
) (gen.DeleteExperimentResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.DeleteExperiment401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	if err := h.service.Delete(ctx, request.ProjectId, request.ExperimentId); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return gen.DeleteExperiment404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("experiment not found")},
			}, nil
		case errors.Is(err, domain.ErrNotDraft):
			return gen.DeleteExperiment409JSONResponse{
				ConflictJSONResponse: gen.ConflictJSONResponse{Message: ptr("experiment must be in draft status")},
			}, nil
		}

		return nil, err
	}

	return gen.DeleteExperiment204Response{}, nil
}

func (h *ExperimentHandler) StartExperiment(
	ctx context.Context,
	request gen.StartExperimentRequestObject,
) (gen.StartExperimentResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.StartExperiment401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	exp, err := h.service.Start(ctx, request.ProjectId, request.ExperimentId)
	if err != nil {
		return handleTransitionError[gen.StartExperimentResponseObject](
			err,
			gen.StartExperiment404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("experiment not found")},
			},
			gen.StartExperiment409JSONResponse{
				ConflictJSONResponse: gen.ConflictJSONResponse{Message: ptr("invalid status transition")},
			},
		)
	}

	return gen.StartExperiment200JSONResponse(toAPIExperiment(exp)), nil
}

func (h *ExperimentHandler) PauseExperiment(
	ctx context.Context,
	request gen.PauseExperimentRequestObject,
) (gen.PauseExperimentResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.PauseExperiment401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	exp, err := h.service.Pause(ctx, request.ProjectId, request.ExperimentId)
	if err != nil {
		return handleTransitionError[gen.PauseExperimentResponseObject](
			err,
			gen.PauseExperiment404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("experiment not found")},
			},
			gen.PauseExperiment409JSONResponse{
				ConflictJSONResponse: gen.ConflictJSONResponse{Message: ptr("invalid status transition")},
			},
		)
	}

	return gen.PauseExperiment200JSONResponse(toAPIExperiment(exp)), nil
}

func (h *ExperimentHandler) ResumeExperiment(
	ctx context.Context,
	request gen.ResumeExperimentRequestObject,
) (gen.ResumeExperimentResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.ResumeExperiment401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	exp, err := h.service.Resume(ctx, request.ProjectId, request.ExperimentId)
	if err != nil {
		return handleTransitionError[gen.ResumeExperimentResponseObject](
			err,
			gen.ResumeExperiment404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("experiment not found")},
			},
			gen.ResumeExperiment409JSONResponse{
				ConflictJSONResponse: gen.ConflictJSONResponse{Message: ptr("invalid status transition")},
			},
		)
	}

	return gen.ResumeExperiment200JSONResponse(toAPIExperiment(exp)), nil
}

func (h *ExperimentHandler) CompleteExperiment(
	ctx context.Context,
	request gen.CompleteExperimentRequestObject,
) (gen.CompleteExperimentResponseObject, error) {
	if _, ok := middleware.UserIDFromContext(ctx); !ok {
		return gen.CompleteExperiment401JSONResponse{
			UnauthorizedJSONResponse: gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")},
		}, nil
	}

	exp, err := h.service.Complete(ctx, request.ProjectId, request.ExperimentId)
	if err != nil {
		return handleTransitionError[gen.CompleteExperimentResponseObject](
			err,
			gen.CompleteExperiment404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("experiment not found")},
			},
			gen.CompleteExperiment409JSONResponse{
				ConflictJSONResponse: gen.ConflictJSONResponse{Message: ptr("invalid status transition")},
			},
		)
	}

	return gen.CompleteExperiment200JSONResponse(toAPIExperiment(exp)), nil
}

func handleTransitionError[R any](err error, notFound, conflict R) (R, error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return notFound, nil
	case errors.Is(err, domain.ErrInvalidTransition):
		return conflict, nil
	}

	var zero R
	return zero, err
}
