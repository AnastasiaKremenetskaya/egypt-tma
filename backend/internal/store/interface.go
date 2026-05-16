package store

import "github.com/anterekhova/egypt-tma/internal/game"

// Store is the persistence layer for rooms.
// Swap memory.go for a Redis implementation without touching game logic.
type Store interface {
	Create(room *game.Room) error
	Get(code string) (*game.Room, error)
	Save(room *game.Room) error
	Delete(code string) error
	FindByPlayer(userID int64) (*game.Room, error)
}
