package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/data"
)

func TestHashAndVerify(t *testing.T) {
	password := "secret123"
	hashed := HashPassword(password)

	if !VerifyPassword(password, hashed) {
		t.Fatal("VerifyPassword should return true for correct password")
	}

	if VerifyPassword("wrong", hashed) {
		t.Fatal("VerifyPassword should return false for wrong password")
	}
}

func newTestStore(t *testing.T) *data.UserStore {
	t.Helper()
	store := data.NewUserStore("")
	if err := store.Load(); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestLoginHandler(t *testing.T) {
	store := newTestStore(t)
	store.Create(&data.User{
		Username:    "admin",
		Password:    HashPassword("admin"),
		DisplayName: "Administrador",
		Role:        "admin",
		HospitalID:  4,
	})

	sessions := NewSessionStore(0)
	handler := LoginHandler(store, sessions)

	body := `{"username":"admin","password":"admin"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bodyReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp loginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if resp.User.Username != "admin" {
		t.Fatalf("expected admin, got %s", resp.User.Username)
	}
	if resp.User.Role != "admin" {
		t.Fatalf("expected admin role, got %s", resp.User.Role)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	store := newTestStore(t)
	store.Create(&data.User{
		Username: "admin",
		Password: HashPassword("admin"),
	})

	sessions := NewSessionStore(0)
	handler := LoginHandler(store, sessions)

	body := `{"username":"admin","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bodyReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLogoutHandler(t *testing.T) {
	sessions := NewSessionStore(0)
	session := sessions.Create(1, "admin", "Admin", "admin", 4)

	handler := LogoutHandler(sessions)

	req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	req.Header.Set("Authorization", "Bearer "+session.Token)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	_, ok := sessions.Get(session.Token)
	if ok {
		t.Fatal("session should be deleted after logout")
	}
}

func TestAuthMiddleware(t *testing.T) {
	sessions := NewSessionStore(0)
	session := sessions.Create(1, "admin", "Admin", "admin", 4)

	middleware := AuthMiddleware(sessions)

	called := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		s := r.Context().Value(sessionKey).(*Session)
		if s.Username != "admin" {
			t.Fatal("expected admin in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+session.Token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Fatal("handler should have been called")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddlewareNoToken(t *testing.T) {
	sessions := NewSessionStore(0)
	middleware := AuthMiddleware(sessions)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called without token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func bodyReader(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}
