package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type userFile struct {
	NextID int    `json:"next_id"`
	Users  []User `json:"users"`
}

type UserStore struct {
	mu       sync.RWMutex
	users    []User
	nextID   int
	filePath string
}

func NewUserStore(filePath string) *UserStore {
	return &UserStore{
		filePath: filePath,
		nextID:   1,
	}
}

func (s *UserStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var f userFile
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	s.users = f.Users
	s.nextID = f.NextID
	return nil
}

func (s *UserStore) Save() error {
	f := userFile{
		NextID: s.nextID,
		Users:  s.users,
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.filePath), 0755); err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

func (s *UserStore) List() ([]User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]User, len(s.users))
	copy(result, s.users)
	return result, nil
}

func (s *UserStore) GetByUsername(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.users {
		if s.users[i].Username == username {
			u := s.users[i]
			return &u, nil
		}
	}
	return nil, fmt.Errorf("user not found: %s", username)
}

func (s *UserStore) GetByID(id int) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.users {
		if s.users[i].ID == id {
			u := s.users[i]
			return &u, nil
		}
	}
	return nil, fmt.Errorf("user not found: id=%d", id)
}

func (s *UserStore) Create(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user.ID = s.nextID
	s.nextID++
	now := time.Now().UTC().Format(time.RFC3339)
	user.CreatedAt = now
	user.UpdatedAt = now

	s.users = append(s.users, *user)
	return s.Save()
}

func (s *UserStore) Update(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.users {
		if s.users[i].ID == user.ID {
			user.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			s.users[i] = *user
			return s.Save()
		}
	}
	return fmt.Errorf("user not found: id=%d", user.ID)
}

func (s *UserStore) SyncCreate(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users = append(s.users, *user)
	if user.ID >= s.nextID {
		s.nextID = user.ID + 1
	}
	return s.Save()
}

func (s *UserStore) SyncUpdate(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.users {
		if s.users[i].ID == user.ID {
			s.users[i] = *user
			return s.Save()
		}
	}
	return fmt.Errorf("user not found: id=%d", user.ID)
}

func (s *UserStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.users {
		if s.users[i].ID == id {
			s.users = append(s.users[:i], s.users[i+1:]...)
			return s.Save()
		}
	}
	return fmt.Errorf("user not found: id=%d", id)
}
