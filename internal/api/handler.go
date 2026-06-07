package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/data"
)

type errorResponse struct {
	Error string `json:"error"`
}

type EntityAPI[T data.Entity] struct {
	Store *data.GenericStore[T]
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (api *EntityAPI[T]) List(w http.ResponseWriter, r *http.Request) {
	items, err := api.Store.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	if items == nil {
		items = make([]T, 0)
	}
	writeJSON(w, http.StatusOK, items)
}

func (api *EntityAPI[T]) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
		return
	}
	item, err := api.Store.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (api *EntityAPI[T]) Create(w http.ResponseWriter, r *http.Request) {
	var item T
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}
	if err := api.Store.Create(item); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (api *EntityAPI[T]) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
		return
	}
	var item T
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}
	if item.GetID() != 0 && item.GetID() != id {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "id mismatch"})
		return
	}
	item.SetID(id)
	if err := api.Store.Update(item); err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
		return
	}
	// Re-fetch to get updated timestamps
	updated, err := api.Store.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (api *EntityAPI[T]) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
		return
	}
	if err := api.Store.Delete(id); err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
