package lock

import (
	"encoding/json"
	"net/http"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/auth"
)

type Handler struct {
	Manager *LockManager
}

func (h *Handler) Acquire(w http.ResponseWriter, r *http.Request) {
	var req LockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	session, _ := r.Context().Value(auth.SessionKey).(*auth.Session)

	resp := h.Manager.Acquire(req.Resource, session.Username, session.DisplayName, req.TTL)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) Release(w http.ResponseWriter, r *http.Request) {
	var req UnlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	session, _ := r.Context().Value(auth.SessionKey).(*auth.Session)

	ok := h.Manager.Release(req.Resource, session.Username)
	writeJSON(w, http.StatusOK, map[string]bool{"released": ok})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	locks := h.Manager.List()
	if locks == nil {
		locks = []LockEntry{}
	}
	writeJSON(w, http.StatusOK, locks)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
