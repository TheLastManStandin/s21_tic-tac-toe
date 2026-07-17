package datasource

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresDatabase() (*PostgresDatabase, error) {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_CONNECTION_STRING"))
	if err != nil {
		return nil, err
	}
	log.Println("Connected to postgres database")
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

func (db *PostgresDatabase) GetUserByName(username string) (User, error) {
	ret := User{}

	err := db.pool.QueryRow(
		context.Background(),
		"SELECT * FROM client WHERE name = $1",
		username,
	).Scan(&ret.UUID, &ret.Username, &ret.HashPassword)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ret, errors.New("user not found")
		}
		return ret, err
	}

	return ret, nil
}

func (db *PostgresDatabase) AddUser(user User) error {
	_, err := db.pool.Exec(
		context.Background(),
		"INSERT INTO client(uuid, name, password) VALUES($1,$2,$3)",
		user.UUID, user.Username, user.HashPassword,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrTooManyRows) {
			return errors.New("user already exists")
		}
		return err
	}
	return nil
}
