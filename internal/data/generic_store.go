package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Entity interface {
	GetID() int
	SetID(int)
	SetCreatedAt(string)
	SetUpdatedAt(string)
	IsActive() bool
	SetDeleted(string)
}

type entityFile[T any] struct {
	NextID int `json:"next_id"`
	Items  []T `json:"items"`
}

type GenericStore[T Entity] struct {
	mu       sync.RWMutex
	items    []T
	nextID   int
	filePath string
}

func NewGenericStore[T Entity](filePath string) *GenericStore[T] {
	return &GenericStore[T]{
		filePath: filePath,
		nextID:   1,
	}
}

func (s *GenericStore[T]) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var f entityFile[T]
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	s.items = f.Items
	s.nextID = f.NextID
	return nil
}

func (s *GenericStore[T]) Save() error {
	f := entityFile[T]{
		NextID: s.nextID,
		Items:  s.items,
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

func (s *GenericStore[T]) List() ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []T
	for _, item := range s.items {
		if item.IsActive() {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *GenericStore[T]) ListAll() ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]T, len(s.items))
	copy(result, s.items)
	return result, nil
}

func (s *GenericStore[T]) GetByID(id int) (T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.GetID() == id {
			return item, nil
		}
	}
	var zero T
	return zero, fmt.Errorf("entity not found: id=%d", id)
}

func (s *GenericStore[T]) Create(item T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item.SetID(s.nextID)
	s.nextID++
	now := time.Now().UTC().Format(time.RFC3339)
	item.SetCreatedAt(now)
	item.SetUpdatedAt(now)

	s.items = append(s.items, item)
	return s.Save()
}

func (s *GenericStore[T]) Update(item T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.items {
		if s.items[i].GetID() == item.GetID() {
			item.SetUpdatedAt(time.Now().UTC().Format(time.RFC3339))
			s.items[i] = item
			return s.Save()
		}
	}
	return fmt.Errorf("entity not found: id=%d", item.GetID())
}

func (s *GenericStore[T]) SyncCreate(item T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	if item.GetID() >= s.nextID {
		s.nextID = item.GetID() + 1
	}
	return s.Save()
}

func (s *GenericStore[T]) SyncUpdate(item T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].GetID() == item.GetID() {
			s.items[i] = item
			return s.Save()
		}
	}
	return fmt.Errorf("entity not found: id=%d", item.GetID())
}

func (s *GenericStore[T]) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.items {
		if s.items[i].GetID() == id {
			if !s.items[i].IsActive() {
				return fmt.Errorf("entity not found: id=%d", id)
			}
			now := time.Now().UTC().Format(time.RFC3339)
			s.items[i].SetDeleted(now)
			s.items[i].SetUpdatedAt(now)
			return s.Save()
		}
	}
	return fmt.Errorf("entity not found: id=%d", id)
}
