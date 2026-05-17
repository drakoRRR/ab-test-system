package project_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domainproject "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/project"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

func TestProjectHandler_CreateProject(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		body       *gen.CreateProjectJSONRequestBody
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.CreateProjectResponseObject)
	}

	validBody := &gen.CreateProjectJSONRequestBody{Name: "My Project"}

	tests := []testCase{
		{
			name: "201 on success",
			ctx:  authedCtx(),
			body: validBody,
			setupMock: func(mh *mockedHandler) {
				withUserOK(mh)
				mh.projSvc.EXPECT().Create(mock.Anything, fixedOrgID, "My Project", "").Return(fixedProject, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateProjectResponseObject) {
				assert.IsType(t, gen.CreateProject201JSONResponse{}, resp)
			},
		},
		{
			name:      "400 on nil body",
			ctx:       authedCtx(),
			body:      nil,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateProjectResponseObject) {
				assert.IsType(t, gen.CreateProject400JSONResponse{}, resp)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			body:      validBody,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateProjectResponseObject) {
				assert.IsType(t, gen.CreateProject401JSONResponse{}, resp)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			body:      validBody,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateProjectResponseObject) {
				assert.IsType(t, gen.CreateProject401JSONResponse{}, resp)
			},
		},
		{
			name:      "401 when user has no organization",
			ctx:       authedCtx(),
			body:      validBody,
			setupMock: withUserNoOrg,
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateProjectResponseObject) {
				assert.IsType(t, gen.CreateProject401JSONResponse{}, resp)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.CreateProject(tc.ctx, gen.CreateProjectRequestObject{Body: tc.body})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestProjectHandler_ListProjects(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.ListProjectsResponseObject)
	}

	tests := []testCase{
		{
			name: "200 with projects",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				withUserOK(mh)
				mh.projSvc.EXPECT().List(mock.Anything, fixedOrgID).Return([]domainproject.Project{fixedProject}, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.ListProjectsResponseObject) {
				list, ok := resp.(gen.ListProjects200JSONResponse)
				assert.True(t, ok)
				assert.Len(t, list, 1)
			},
		},
		{
			name:      "401 when user has no organization",
			ctx:       authedCtx(),
			setupMock: withUserNoOrg,
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.ListProjectsResponseObject) {
				assert.IsType(t, gen.ListProjects401JSONResponse{}, resp)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.ListProjects(tc.ctx, gen.ListProjectsRequestObject{})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestProjectHandler_GetProject(t *testing.T) {
	type testCase struct {
		name       string
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.GetProjectResponseObject)
	}

	projectParam := fixedProjectID

	tests := []testCase{
		{
			name: "200 on success",
			setupMock: func(mh *mockedHandler) {
				withUserOK(mh)
				mh.projSvc.EXPECT().GetByID(mock.Anything, fixedOrgID, fixedProjectID).Return(fixedProject, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.GetProjectResponseObject) {
				assert.IsType(t, gen.GetProject200JSONResponse{}, resp)
			},
		},
		{
			name: "404 when project not found",
			setupMock: func(mh *mockedHandler) {
				withUserOK(mh)
				mh.projSvc.EXPECT().GetByID(mock.Anything, fixedOrgID, fixedProjectID).
					Return(domainproject.Project{}, domainproject.ErrNotFound)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.GetProjectResponseObject) {
				assert.IsType(t, gen.GetProject404JSONResponse{}, resp)
			},
		},
		{
			name: "service error bubbles up as 500",
			setupMock: func(mh *mockedHandler) {
				withUserOK(mh)
				mh.projSvc.EXPECT().GetByID(mock.Anything, fixedOrgID, fixedProjectID).
					Return(domainproject.Project{}, errors.New("db error"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.GetProject(authedCtx(), gen.GetProjectRequestObject{ProjectId: projectParam})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestProjectHandler_UpdateProject(t *testing.T) {
	type testCase struct {
		name       string
		body       *gen.UpdateProjectJSONRequestBody
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.UpdateProjectResponseObject)
	}

	projectParam := fixedProjectID
	validBody := &gen.UpdateProjectJSONRequestBody{Name: ptr("Updated")}

	tests := []testCase{
		{
			name: "200 on success",
			body: validBody,
			setupMock: func(mh *mockedHandler) {
				withUserOK(mh)
				mh.projSvc.EXPECT().
					Update(mock.Anything, domainproject.UpdateParams{
						OrgID:     fixedOrgID,
						ProjectID: fixedProjectID,
						Name:      ptr("Updated"),
					}).
					Return(fixedProject, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.UpdateProjectResponseObject) {
				assert.IsType(t, gen.UpdateProject200JSONResponse{}, resp)
			},
		},
		{
			name:      "400 on nil body",
			body:      nil,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.UpdateProjectResponseObject) {
				assert.IsType(t, gen.UpdateProject400JSONResponse{}, resp)
			},
		},
		{
			name: "404 when project not found",
			body: validBody,
			setupMock: func(mh *mockedHandler) {
				withUserOK(mh)
				mh.projSvc.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(domainproject.Project{}, domainproject.ErrNotFound)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.UpdateProjectResponseObject) {
				assert.IsType(t, gen.UpdateProject404JSONResponse{}, resp)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.UpdateProject(authedCtx(), gen.UpdateProjectRequestObject{
				ProjectId: projectParam,
				Body:      tc.body,
			})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestProjectHandler_DeleteProject(t *testing.T) {
	type testCase struct {
		name       string
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.DeleteProjectResponseObject)
	}

	projectParam := fixedProjectID

	tests := []testCase{
		{
			name: "204 on success",
			setupMock: func(mh *mockedHandler) {
				withUserOK(mh)
				mh.projSvc.EXPECT().Delete(mock.Anything, fixedOrgID, fixedProjectID).Return(nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.DeleteProjectResponseObject) {
				assert.IsType(t, gen.DeleteProject204Response{}, resp)
			},
		},
		{
			name: "404 when project not found",
			setupMock: func(mh *mockedHandler) {
				withUserOK(mh)
				mh.projSvc.EXPECT().Delete(mock.Anything, fixedOrgID, fixedProjectID).Return(domainproject.ErrNotFound)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.DeleteProjectResponseObject) {
				assert.IsType(t, gen.DeleteProject404JSONResponse{}, resp)
			},
		},
		{
			name: "service error bubbles up as 500",
			setupMock: func(mh *mockedHandler) {
				withUserOK(mh)
				mh.projSvc.EXPECT().Delete(mock.Anything, fixedOrgID, fixedProjectID).Return(errors.New("db error"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.DeleteProject(authedCtx(), gen.DeleteProjectRequestObject{ProjectId: projectParam})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}
