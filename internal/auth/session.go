package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Session struct {
	Token       string
	UserID      int
	Username    string
	DisplayName string
	Role        string
	HospitalID  int
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	timeout  time.Duration
}

func NewSessionStore(timeout time.Duration) *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
		timeout:  timeout,
	}
}

func (s *SessionStore) Create(userID int, username, displayName, role string, hospitalID int) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	token := generateToken()
	now := time.Now()
	session := &Session{
		Token:       token,
		UserID:      userID,
		Username:    username,
		DisplayName: displayName,
		Role:        role,
		HospitalID:  hospitalID,
		CreatedAt:   now,
		ExpiresAt:   now.Add(s.timeout),
	}
	s.sessions[token] = session
	return session
}

func (s *SessionStore) Get(token string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[token]
	if !ok {
		return nil, false
	}
	if s.timeout > 0 && time.Now().After(session.ExpiresAt) {
		return nil, false
	}
	return session, true
}

func (s *SessionStore) Delete(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
