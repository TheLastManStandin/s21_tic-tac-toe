package datasource

import (
	"context"
	"log"
	"os"
	"tickTackToe/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

type PostgresDatabase struct {
	pool *pgxpool.Pool
}

func NewPostgresDatabase(lc fx.Lifecycle) *PostgresDatabase {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_CONNECTION_STRING"))
	if err != nil {
		panic(err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		panic(err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})
	log.Println("Connected to postgres database")
	return &PostgresDatabase{pool: pool}
}

func (db *PostgresDatabase) SaveGame(ctx context.Context, game domain.GameInstance, UUID string) error {
	table := make([][]int32, 3)
	for i := 0; i < 3; i++ {
		table[i] = make([]int32, 3)
		for j := 0; j < 3; j++ {
			table[i][j] = int32(game.Table[i][j])
		}
	}

	_, err := db.pool.Exec(
		context.Background(),
		"INSERT INTO games(uuid, gametable, gamestate) VALUES($1,$2,$3) "+
			"ON CONFLICT (uuid) DO UPDATE SET gametable = $2, gamestate = $3",
		game.UUID, table, int(game.State),
	)
	if err != nil {
		//todo обработать если база не отвечает
		return err
	}
	return nil
}

func FindByID(ctx context.Context, UUID string) (GameRow, error) {

}
