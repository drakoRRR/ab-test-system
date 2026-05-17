package sdk_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domainexp "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
	domainflag "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/flag"
	domainsdk "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/sdk"
)

func TestService_GetConfig(t *testing.T) {
	type testCase struct {
		name      string
		setupMock func(ms *mockedService)
		wantErr   bool
		assertRes func(t *testing.T, cfg domainsdk.Config)
	}

	tests := []testCase{
		{
			name: "returns config with flags and only running experiments",
			setupMock: func(ms *mockedService) {
				ms.flags.EXPECT().
					List(mock.Anything, projectID).
					Return([]domainflag.Flag{activeFlag}, nil)
				ms.experiments.EXPECT().
					List(mock.Anything, projectID).
					Return([]domainexp.Experiment{runningExperiment, draftExperiment, pausedExperiment}, nil)
			},
			assertRes: func(t *testing.T, cfg domainsdk.Config) {
				assert.Equal(t, projectID, cfg.ProjectID)
				require.Len(t, cfg.Flags, 1)
				assert.Equal(t, activeFlag.Key, cfg.Flags[0].Key)
				require.Len(t, cfg.Experiments, 1)
				assert.Equal(t, runningExperiment.Key, cfg.Experiments[0].Key)
				assert.Equal(t, domainexp.StatusRunning, cfg.Experiments[0].Status)
			},
		},
		{
			name: "returns empty slices when no flags or experiments",
			setupMock: func(ms *mockedService) {
				ms.flags.EXPECT().
					List(mock.Anything, projectID).
					Return([]domainflag.Flag{}, nil)
				ms.experiments.EXPECT().
					List(mock.Anything, projectID).
					Return([]domainexp.Experiment{}, nil)
			},
			assertRes: func(t *testing.T, cfg domainsdk.Config) {
				assert.Equal(t, projectID, cfg.ProjectID)
				assert.Empty(t, cfg.Flags)
				assert.Empty(t, cfg.Experiments)
			},
		},
		{
			name: "filters out draft and paused experiments",
			setupMock: func(ms *mockedService) {
				ms.flags.EXPECT().
					List(mock.Anything, projectID).
					Return([]domainflag.Flag{}, nil)
				ms.experiments.EXPECT().
					List(mock.Anything, projectID).
					Return([]domainexp.Experiment{draftExperiment, pausedExperiment}, nil)
			},
			assertRes: func(t *testing.T, cfg domainsdk.Config) {
				assert.Empty(t, cfg.Experiments)
			},
		},
		{
			name: "propagates flag lister error",
			setupMock: func(ms *mockedService) {
				ms.flags.EXPECT().
					List(mock.Anything, projectID).
					Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "propagates experiment lister error",
			setupMock: func(ms *mockedService) {
				ms.flags.EXPECT().
					List(mock.Anything, projectID).
					Return([]domainflag.Flag{activeFlag}, nil)
				ms.experiments.EXPECT().
					List(mock.Anything, projectID).
					Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setupMock(ms)

			cfg, err := ms.GetConfig(context.Background(), projectID)

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.assertRes != nil {
				tc.assertRes(t, cfg)
			}
		})
	}
}
