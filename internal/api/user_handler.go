package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/auth"
	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/data"
)

type UserAPI struct {
	Store   *data.UserStore
	OnWrite func(action, model string, id int, data json.RawMessage)
}

type userResponse struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	HospitalID  int    `json:"hospital_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type createUserRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	HospitalID  int    `json:"hospital_id"`
}

type updateUserRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password,omitempty"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	HospitalID  int    `json:"hospital_id"`
}

func toUserResponse(u *data.User) userResponse {
	return userResponse{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		HospitalID:  u.HospitalID,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

func (api *UserAPI) List(w http.ResponseWriter, r *http.Request) {
	users, err := api.Store.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	resp := make([]userResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, toUserResponse(&u))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (api *UserAPI) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
		return
	}
	u, err := api.Store.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(u))
}

func (api *UserAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}
	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "username and password required"})
		return
	}
	user := &data.User{
		Username:    req.Username,
		Password:    auth.HashPassword(req.Password),
		DisplayName: req.DisplayName,
		Role:        req.Role,
		HospitalID:  req.HospitalID,
	}
	if user.Role == "" {
		user.Role = "doctor"
	}
	if user.HospitalID == 0 {
		user.HospitalID = 1
	}
	if err := api.Store.Create(user); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	if api.OnWrite != nil {
		d, _ := json.Marshal(user)
		api.OnWrite("create", "usuario", user.ID, d)
	}
	writeJSON(w, http.StatusCreated, toUserResponse(user))
}

func (api *UserAPI) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
		return
	}

	existing, err := api.Store.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	existing.Username = req.Username
	existing.DisplayName = req.DisplayName
	existing.Role = req.Role
	existing.HospitalID = req.HospitalID
	if req.Password != "" {
		existing.Password = auth.HashPassword(req.Password)
	}

	if err := api.Store.Update(existing); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	if api.OnWrite != nil {
		d, _ := json.Marshal(existing)
		api.OnWrite("update", "usuario", id, d)
	}
	writeJSON(w, http.StatusOK, toUserResponse(existing))
}

func (api *UserAPI) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid id"})
		return
	}
	// Prevent deleting yourself
	session, _ := r.Context().Value(auth.SessionKey).(*auth.Session)
	if session != nil && session.UserID == id {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "cannot delete yourself"})
		return
	}
	if err := api.Store.Delete(id); err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
		return
	}
	if api.OnWrite != nil {
		api.OnWrite("delete", "usuario", id, nil)
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
