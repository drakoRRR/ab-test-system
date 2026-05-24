package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestHandler(sdk *MocksdkClient) *Handler {
	return &Handler{sdk: sdk, cfg: Config{ExperimentKey: "test-exp", FlagKey: "test-flag"}}
}

func TestHandleVisit(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		setupMock   func(*MocksdkClient)
		wantStatus  int
		wantVariant string
	}{
		{
			name:  "returns assigned variant",
			query: "user_id=u1&experiment_key=checkout-btn",
			setupMock: func(m *MocksdkClient) {
				m.EXPECT().GetVariant("checkout-btn", "u1").Return("treatment")
			},
			wantStatus:  http.StatusOK,
			wantVariant: "treatment",
		},
		{
			name:  "empty variant when user outside traffic allocation",
			query: "user_id=u1&experiment_key=checkout-btn",
			setupMock: func(m *MocksdkClient) {
				m.EXPECT().GetVariant("checkout-btn", "u1").Return("")
			},
			wantStatus:  http.StatusOK,
			wantVariant: "",
		},
		{
			name:  "falls back to cfg.ExperimentKey when param absent",
			query: "user_id=u1",
			setupMock: func(m *MocksdkClient) {
				m.EXPECT().GetVariant("test-exp", "u1").Return("control")
			},
			wantStatus:  http.StatusOK,
			wantVariant: "control",
		},
		{
			name:       "missing user_id returns 400",
			query:      "",
			setupMock:  func(_ *MocksdkClient) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMocksdkClient(t)
			tt.setupMock(m)

			req := httptest.NewRequest(http.MethodGet, "/visit?"+tt.query, nil)
			w := httptest.NewRecorder()
			newTestHandler(m).handleVisit(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp visitResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				assert.Equal(t, tt.wantVariant, resp.Variant)
			}
		})
	}
}

func TestHandleConvert(t *testing.T) {
	tests := []struct {
		name       string
		body       convertRequest
		wantStatus int
		wantEvent  string
	}{
		{
			name:       "tracks named event",
			body:       convertRequest{UserID: "u1", Event: "purchase", Value: 1.0},
			wantStatus: http.StatusOK,
			wantEvent:  "purchase",
		},
		{
			name:       "defaults event name to purchase",
			body:       convertRequest{UserID: "u1"},
			wantStatus: http.StatusOK,
			wantEvent:  "purchase",
		},
		{
			name:       "missing user_id returns 400",
			body:       convertRequest{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMocksdkClient(t)
			if tt.wantStatus == http.StatusOK {
				m.EXPECT().Track(tt.wantEvent, "u1", mock.Anything).Return()
			}

			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/convert", bytes.NewReader(b))
			w := httptest.NewRecorder()
			newTestHandler(m).handleConvert(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestHandleFlag(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		setupMock   func(*MocksdkClient)
		wantStatus  int
		wantEnabled bool
	}{
		{
			name:  "returns enabled=true",
			query: "user_id=u1&flag_key=my-flag",
			setupMock: func(m *MocksdkClient) {
				m.EXPECT().IsEnabled("my-flag", "u1").Return(true)
			},
			wantStatus:  http.StatusOK,
			wantEnabled: true,
		},
		{
			name:  "falls back to cfg.FlagKey when param absent",
			query: "user_id=u1",
			setupMock: func(m *MocksdkClient) {
				m.EXPECT().IsEnabled("test-flag", "u1").Return(false)
			},
			wantStatus:  http.StatusOK,
			wantEnabled: false,
		},
		{
			name:       "missing user_id returns 400",
			query:      "flag_key=my-flag",
			setupMock:  func(_ *MocksdkClient) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMocksdkClient(t)
			tt.setupMock(m)

			req := httptest.NewRequest(http.MethodGet, "/flag?"+tt.query, nil)
			w := httptest.NewRecorder()
			newTestHandler(m).handleFlag(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp flagResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				assert.Equal(t, tt.wantEnabled, resp.Enabled)
			}
		})
	}
}

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	newTestHandler(NewMocksdkClient(t)).handleHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}
