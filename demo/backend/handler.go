package main

import (
	"encoding/json"
	"net/http"
)

type sdkClient interface {
	IsEnabled(flagKey, userID string) bool
	GetVariant(experimentKey, userID string) string
	Track(eventName, userID string, value float64)
}

type Handler struct {
	sdk sdkClient
	cfg Config
}

type visitResponse struct {
	Variant string `json:"variant"`
}

type convertRequest struct {
	UserID string  `json:"user_id"`
	Event  string  `json:"event"`
	Value  float64 `json:"value"`
}

type flagResponse struct {
	Enabled bool `json:"enabled"`
}

func (h *Handler) handleVisit(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, `{"error":"user_id required"}`, http.StatusBadRequest)
		return
	}
	expKey := r.URL.Query().Get("experiment_key")
	if expKey == "" {
		expKey = h.cfg.ExperimentKey
	}
	respond(w, visitResponse{Variant: h.sdk.GetVariant(expKey, userID)})
}

func (h *Handler) handleConvert(w http.ResponseWriter, r *http.Request) {
	var req convertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == "" {
		http.Error(w, `{"error":"user_id required"}`, http.StatusBadRequest)
		return
	}
	event := req.Event
	if event == "" {
		event = "purchase"
	}
	h.sdk.Track(event, req.UserID, req.Value)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleFlag(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, `{"error":"user_id required"}`, http.StatusBadRequest)
		return
	}
	flagKey := r.URL.Query().Get("flag_key")
	if flagKey == "" {
		flagKey = h.cfg.FlagKey
	}
	respond(w, flagResponse{Enabled: h.sdk.IsEnabled(flagKey, userID)})
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	respond(w, map[string]string{"status": "ok"})
}

func respond(w http.ResponseWriter, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}
