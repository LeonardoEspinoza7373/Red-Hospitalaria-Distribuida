package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/data"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string   `json:"token"`
	User  userInfo `json:"user"`
}

type userInfo struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	HospitalID  int    `json:"hospital_id"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type contextKey string

const sessionKey contextKey = "session"

func LoginHandler(store *data.UserStore, sessions *SessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
			return
		}

		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
			return
		}

		user, err := store.GetByUsername(req.Username)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
			return
		}

		if !VerifyPassword(req.Password, user.Password) {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
			return
		}

		session := sessions.Create(user.ID, user.Username, user.DisplayName, user.Role, user.HospitalID)

		writeJSON(w, http.StatusOK, loginResponse{
			Token: session.Token,
			User: userInfo{
				ID:          user.ID,
				Username:    user.Username,
				DisplayName: user.DisplayName,
				Role:        user.Role,
				HospitalID:  user.HospitalID,
			},
		})
	}
}

func LogoutHandler(sessions *SessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
			return
		}

		token := extractToken(r)
		if token != "" {
			sessions.Delete(token)
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
	}
}

func MeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := r.Context().Value(sessionKey).(*Session)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, userInfo{
			ID:          session.UserID,
			Username:    session.Username,
			DisplayName: session.DisplayName,
			Role:        session.Role,
			HospitalID:  session.HospitalID,
		})
	}
}

func AuthMiddleware(sessions *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
				return
			}
			session, ok := sessions.Get(token)
			if !ok {
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
				return
			}
			ctx := context.WithValue(r.Context(), sessionKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
