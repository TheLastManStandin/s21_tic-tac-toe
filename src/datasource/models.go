package datasource

import (
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database interface {
	SaveGame(instance GameInstance, UUID string) error
	LoadGame(UUID string) (GameInstance, error)
	Close() error
}

type MapDatabase struct {
	Games sync.Map
}

type PostgresDatabase struct {
	pool *pgxpool.Pool
}

type GameInstance struct {
	Table [3][3]int
}
