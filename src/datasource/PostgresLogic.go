package datasource

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresDatabase() (*PostgresDatabase, error) {
	//todo убрать хардкод секретов бд
	pool, err := pgxpool.New(context.Background(), getConnString())
	if err != nil {
		return nil, err
	}
	return &PostgresDatabase{pool: pool}, nil
}

func (db *PostgresDatabase) SaveGame(game GameInstance, UUID string) error {
	table := make([][]int32, 3)
	for i := 0; i < 3; i++ {
		table[i] = make([]int32, 3)
		for j := 0; j < 3; j++ {
			table[i][j] = int32(game.Table[i][j])
		}
	}

	_, err := db.pool.Exec(
		context.Background(),
		"INSERT INTO games(uuid, gametable) VALUES($1,$2) ON CONFLICT (uuid) DO UPDATE SET gametable = $2",
		UUID, table,
	)
	if err != nil {
		//todo обработать если база не отвечает
		return err
	}
	return nil
}

func (db *PostgresDatabase) LoadGame(UUID string) (GameInstance, error) {
	ret := GameInstance{}

	var raw [][]int32
	err := db.pool.QueryRow(
		context.Background(),
		"SELECT gametable FROM games WHERE uuid = $1",
		UUID,
	).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ret, errors.New("game not found")
		}
		return ret, err
	}

	if len(raw) != 3 || len(raw[0]) != 3 {
		return ret, errors.New("invalid table dimensions in database")
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			ret.Table[i][j] = int(raw[i][j])
		}
	}

	return ret, nil
}

func (db *PostgresDatabase) Close() error {
	db.pool.Close()
	return nil
}

func getConnString() string {
	ret := strings.Builder{}
	connString := os.Getenv("DATABASE_CONNSTRING")
	ret.WriteString(connString)

	return ret.String()
}
