package api

import (
	"encoding/json"
	"net/http"
)

type BullyHandler struct {
	IsEnabled func() bool
	SetEnabled func(bool)
}

type bullyStatus struct {
	Enabled bool `json:"enabled"`
}

func (h *BullyHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, bullyStatus{Enabled: h.IsEnabled()})
}

func (h *BullyHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	var req bullyStatus
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}
	h.SetEnabled(req.Enabled)
	writeJSON(w, http.StatusOK, bullyStatus{Enabled: req.Enabled})
}
