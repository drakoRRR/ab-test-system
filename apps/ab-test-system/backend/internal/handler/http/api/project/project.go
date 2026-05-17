package project

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domainproject "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/project"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"
)

func (h *ProjectHandler) currentOrgID(ctx context.Context) (uuid.UUID, *gen.UnauthorizedJSONResponse) {
	uid, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return uuid.UUID{}, &gen.UnauthorizedJSONResponse{Message: ptr("unauthorized")}
	}

	u, err := h.users.GetCurrentUser(ctx, uid)
	if err != nil {
		return uuid.UUID{}, &gen.UnauthorizedJSONResponse{Message: ptr("user not found")}
	}

	if u.OrgID == nil {
		return uuid.UUID{}, &gen.UnauthorizedJSONResponse{Message: ptr("user has no organization")}
	}

	return *u.OrgID, nil
}

func (h *ProjectHandler) CreateProject(
	ctx context.Context,
	request gen.CreateProjectRequestObject,
) (gen.CreateProjectResponseObject, error) {
	if request.Body == nil {
		return gen.CreateProject400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: ptr("request body is required")},
		}, nil
	}

	orgID, unauthorized := h.currentOrgID(ctx)
	if unauthorized != nil {
		return gen.CreateProject401JSONResponse{UnauthorizedJSONResponse: *unauthorized}, nil
	}

	description := ""
	if request.Body.Description != nil {
		description = *request.Body.Description
	}

	p, err := h.projects.Create(ctx, orgID, request.Body.Name, description)
	if err != nil {
		return nil, err
	}

	return gen.CreateProject201JSONResponse(toAPIProject(p)), nil
}

func (h *ProjectHandler) ListProjects(
	ctx context.Context,
	_ gen.ListProjectsRequestObject,
) (gen.ListProjectsResponseObject, error) {
	orgID, unauthorized := h.currentOrgID(ctx)
	if unauthorized != nil {
		return gen.ListProjects401JSONResponse{UnauthorizedJSONResponse: *unauthorized}, nil
	}

	projects, err := h.projects.List(ctx, orgID)
	if err != nil {
		return nil, err
	}

	resp := make(gen.ListProjects200JSONResponse, len(projects))
	for i, p := range projects {
		resp[i] = toAPIProject(p)
	}

	return resp, nil
}

func (h *ProjectHandler) GetProject(
	ctx context.Context,
	request gen.GetProjectRequestObject,
) (gen.GetProjectResponseObject, error) {
	orgID, unauthorized := h.currentOrgID(ctx)
	if unauthorized != nil {
		return gen.GetProject401JSONResponse{UnauthorizedJSONResponse: *unauthorized}, nil
	}

	p, err := h.projects.GetByID(ctx, orgID, uuid.UUID(request.ProjectId))
	if err != nil {
		if errors.Is(err, domainproject.ErrNotFound) {
			return gen.GetProject404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("project not found")},
			}, nil
		}

		return nil, err
	}

	return gen.GetProject200JSONResponse(toAPIProject(p)), nil
}

func (h *ProjectHandler) UpdateProject(
	ctx context.Context,
	request gen.UpdateProjectRequestObject,
) (gen.UpdateProjectResponseObject, error) {
	if request.Body == nil {
		return gen.UpdateProject400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: ptr("request body is required")},
		}, nil
	}

	orgID, unauthorized := h.currentOrgID(ctx)
	if unauthorized != nil {
		return gen.UpdateProject401JSONResponse{UnauthorizedJSONResponse: *unauthorized}, nil
	}

	p, err := h.projects.Update(ctx, orgID, uuid.UUID(request.ProjectId), request.Body.Name, request.Body.Description)
	if err != nil {
		if errors.Is(err, domainproject.ErrNotFound) {
			return gen.UpdateProject404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("project not found")},
			}, nil
		}

		return nil, err
	}

	return gen.UpdateProject200JSONResponse(toAPIProject(p)), nil
}

func (h *ProjectHandler) DeleteProject(
	ctx context.Context,
	request gen.DeleteProjectRequestObject,
) (gen.DeleteProjectResponseObject, error) {
	orgID, unauthorized := h.currentOrgID(ctx)
	if unauthorized != nil {
		return gen.DeleteProject401JSONResponse{UnauthorizedJSONResponse: *unauthorized}, nil
	}

	err := h.projects.Delete(ctx, orgID, uuid.UUID(request.ProjectId))
	if err != nil {
		if errors.Is(err, domainproject.ErrNotFound) {
			return gen.DeleteProject404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: ptr("project not found")},
			}, nil
		}

		return nil, err
	}

	return gen.DeleteProject204Response{}, nil
}
