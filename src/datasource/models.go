package datasource

import (
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database interface {
	SaveGame(instance GameInstance, UUID string) error
	LoadGame(UUID string) (GameInstance, error)
	Close() error
	GetUserByName(username string) (User, error)
	AddUser(user User) error
}

type MapDatabase struct {
	Games sync.Map
	Users sync.Map
}

type PostgresDatabase struct {
	pool *pgxpool.Pool
}

type GameInstance struct {
	Table [3][3]int
}

type User struct {
	UUID         string `db:"uuid"`
	Username     string `db:"name"`
	HashPassword []byte `db:"password"`
}
