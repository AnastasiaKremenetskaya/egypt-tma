package store

import (
	"fmt"
	"sync"

	"github.com/anterekhova/egypt-tma/internal/game"
)

type MemoryStore struct {
	mu    sync.RWMutex
	rooms map[string]*game.Room
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{rooms: make(map[string]*game.Room)}
}

func (s *MemoryStore) Create(room *game.Room) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.rooms[room.Code]; exists {
		return fmt.Errorf("room %s already exists", room.Code)
	}
	s.rooms[room.Code] = room
	return nil
}

func (s *MemoryStore) Get(code string) (*game.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rooms[code]
	if !ok {
		return nil, fmt.Errorf("room %s not found", code)
	}
	return r, nil
}

func (s *MemoryStore) Save(room *game.Room) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rooms[room.Code] = room
	return nil
}

func (s *MemoryStore) Delete(code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rooms, code)
	return nil
}

func (s *MemoryStore) FindByPlayer(userID int64) (*game.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.rooms {
		for _, p := range r.Players {
			if p.UserID == userID {
				return r, nil
			}
		}
	}
	return nil, fmt.Errorf("no room for user %d", userID)
}
