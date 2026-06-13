package lock

import (
	"sync"
	"time"
)

type LockEntry struct {
	Resource  string    `json:"resource"`
	HolderID  int       `json:"holder_id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	ExpiresAt time.Time `json:"expires_at"`
}

type LockRequest struct {
	Resource string `json:"resource"`
	TTL      int    `json:"ttl,omitempty"`
}

type LockResponse struct {
	Acquired  bool      `json:"acquired"`
	HolderID  int       `json:"holder_id,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	UserName  string    `json:"user_name,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

type UnlockRequest struct {
	Resource string `json:"resource"`
}

type LockManager struct {
	mu     sync.Mutex
	locks  map[string]*LockEntry
	ttl    time.Duration
	nodeID int
	stopCh chan struct{}
}

func NewLockManager(nodeID int) *LockManager {
	lm := &LockManager{
		locks:  make(map[string]*LockEntry),
		ttl:    60 * time.Second,
		nodeID: nodeID,
		stopCh: make(chan struct{}),
	}
	go lm.cleanupLoop()
	return lm
}

func (lm *LockManager) Stop() {
	close(lm.stopCh)
}

func (lm *LockManager) Acquire(resource string, userID, userName string, ttlSeconds int) LockResponse {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if ttlSeconds <= 0 || ttlSeconds > 300 {
		ttlSeconds = 60
	}

	existing, exists := lm.locks[resource]
	now := time.Now()

	if exists && now.After(existing.ExpiresAt) {
		delete(lm.locks, resource)
		exists = false
	}

	if exists {
		if existing.UserID == userID {
			existing.ExpiresAt = now.Add(time.Duration(ttlSeconds) * time.Second)
			existing.UserName = userName
			return LockResponse{
				Acquired:  true,
				HolderID:  existing.HolderID,
				UserID:    existing.UserID,
				UserName:  existing.UserName,
				ExpiresAt: existing.ExpiresAt,
			}
		}
		return LockResponse{
			Acquired:  false,
			HolderID:  existing.HolderID,
			UserID:    existing.UserID,
			UserName:  existing.UserName,
			ExpiresAt: existing.ExpiresAt,
		}
	}

	entry := &LockEntry{
		Resource:  resource,
		HolderID:  lm.nodeID,
		UserID:    userID,
		UserName:  userName,
		ExpiresAt: now.Add(time.Duration(ttlSeconds) * time.Second),
	}
	lm.locks[resource] = entry

	return LockResponse{
		Acquired:  true,
		HolderID:  lm.nodeID,
		UserID:    userID,
		UserName:  userName,
		ExpiresAt: entry.ExpiresAt,
	}
}

func (lm *LockManager) Release(resource, userID string) bool {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	existing, exists := lm.locks[resource]
	if exists && time.Now().After(existing.ExpiresAt) {
		delete(lm.locks, resource)
		return false
	}
	if !exists || existing.UserID != userID {
		return false
	}

	delete(lm.locks, resource)
	return true
}

func (lm *LockManager) List() []LockEntry {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	now := time.Now()
	result := make([]LockEntry, 0, len(lm.locks))
	for _, entry := range lm.locks {
		if now.Before(entry.ExpiresAt) {
			result = append(result, *entry)
		}
	}
	return result
}

func (lm *LockManager) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			lm.mu.Lock()
			now := time.Now()
			for key, entry := range lm.locks {
				if now.After(entry.ExpiresAt) {
					delete(lm.locks, key)
				}
			}
			lm.mu.Unlock()
		case <-lm.stopCh:
			return
		}
	}
}
