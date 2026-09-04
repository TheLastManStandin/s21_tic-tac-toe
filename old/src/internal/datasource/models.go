package datasource

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database interface {
	SaveGame(instance GameInstance, UUID string) error
	LoadGame(UUID string) (GameInstance, error)
	Close() error
	GetUserByName(username string) (User, error)
	AddUser(user User) error
}

type PostgresDatabase struct {
	pool *pgxpool.Pool
}

type GameState int

// todo добавить енумерацию в sql
const (
	WaitingForPlayers GameState = iota
	FirstPlayerTurn
	SecondPlayerTurn
	FirstPlayerWon
	SecondPlayerWon
	Draw
)

type GameInstance struct {
	Table [3][3]int
	State GameState
}

type User struct {
	UUID         string `db:"uuid"`
	Username     string `db:"name"`
	HashPassword []byte `db:"password"`
}
