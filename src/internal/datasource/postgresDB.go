package datasource

import (
	"context"
	"errors"
	"log"
	"os"
	"tickTackToe/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func (db *PostgresDatabase) SaveGame(ctx context.Context, game domain.GameInstance) error {
	table := make([][]int32, 3)
	for i := 0; i < 3; i++ {
		table[i] = make([]int32, 3)
		for j := 0; j < 3; j++ {
			table[i][j] = int32(game.Table[i][j])
		}
	}

	insertOperation := func() error {
		_, err := db.pool.Exec(
			ctx,
			"INSERT INTO games(uuid, gametable, gamestate, minimax) VALUES($1,$2,$3,$4) "+
				"ON CONFLICT (uuid) DO UPDATE SET gametable = $2, gamestate = $3, minimax = $4",
			game.UUID, table, int(game.State), game.Minimax,
		)
		return err
	}

	return BackoffRetry(ctx, maxAttempts, baseDelay, insertOperation)
}

func (db *PostgresDatabase) FindGameByID(ctx context.Context, UUID string) (domain.GameInstance, error) {
	var raw [][]int32
	var ret domain.GameInstance
	//var logger zap.Logger = ctx.Value("logger").(zap.Logger)

	var queryErr error
	findOperation := func() error {
		queryErr = db.pool.QueryRow(
			ctx,
			`SELECT 
				gametable, 
				gamestate, 
				minimax, 
				firstplayer, 
            	COALESCE(secondplayer::text, '') 
			FROM games WHERE uuid = $1`,
			UUID,
		).Scan(&raw, &ret.State, &ret.Minimax, &ret.FirstPlayerUUID, &ret.SecondPlayerUUID)

		if errors.Is(queryErr, pgx.ErrNoRows) {
			return nil
		}
		return queryErr
	}

	if err := BackoffRetry(ctx, maxAttempts, baseDelay, findOperation); err != nil {
		return ret, err
	}
	if errors.Is(queryErr, pgx.ErrNoRows) {
		return ret, DatabaseGameNotFount{}
	}

	ret.UUID = UUID
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

func (db *PostgresDatabase) InitGame(
	ctx context.Context,
	gameUUID string,
	playerUUID string,
	minimax bool,
) error {
	table := make([][]int32, 3)
	for i := 0; i < 3; i++ {
		table[i] = make([]int32, 3)
		for j := 0; j < 3; j++ {
			table[i][j] = 0
		}
	}

	insertOperation := func() error {
		_, err := db.pool.Exec(
			ctx,
			`INSERT INTO games(
				uuid,
				gametable,
				gamestate,
				minimax,
				firstplayer)
			VALUES($1,$2,$3,$4,$5) 
			ON CONFLICT (uuid) DO NOTHING`,
			gameUUID, table, 0, minimax, playerUUID,
		)
		return err
	}

	return BackoffRetry(ctx, maxAttempts, baseDelay, insertOperation)
}

func (db *PostgresDatabase) ConnectToGame(
	ctx context.Context,
	gameUUID string,
	playerUUID string,
) error {
	connectOperation := func() error {
		result, err := db.pool.Exec(
			ctx,
			`
			UPDATE games
			SET secondplayer = $1,
				gamestate = $2
			WHERE uuid = $3
			  AND minimax = FALSE
			  AND gamestate = $4
			  AND secondplayer IS NULL
			  AND firstplayer IS NOT NULL
			  AND firstplayer <> $1
			`,
			playerUUID,
			domain.FirstPlayerTurn,
			gameUUID,
			domain.WaitingForPlayers,
		)
		if err != nil {
			return err
		}

		if result.RowsAffected() == 0 {
			return domain.ErrCannotConnectToGame
		}

		return nil
	}

	return BackoffRetry(ctx, maxAttempts, baseDelay, connectOperation)
}

func (db *PostgresDatabase) GetUser(ctx context.Context, username string) (domain.UserInstance, error) {
	userRow := UserRow{}
	var queryErr error

	getOperation := func() error {
		userRow = UserRow{}
		queryErr = db.pool.QueryRow(
			ctx,
			"SELECT uuid, name, password FROM client WHERE name = $1",
			username,
		).Scan(&userRow.UUID, &userRow.Username, &userRow.Password)

		if errors.Is(queryErr, pgx.ErrNoRows) {
			return nil
		}
		return queryErr
	}

	if err := BackoffRetry(ctx, maxAttempts, baseDelay, getOperation); err != nil {
		return domain.UserInstance{}, err
	}
	if errors.Is(queryErr, pgx.ErrNoRows) {
		return domain.UserInstance{}, queryErr
	}

	return domain.UserInstance{
		UUID:     userRow.UUID,
		Username: userRow.Username,
		Password: userRow.Password,
	}, nil
}

func (db *PostgresDatabase) InsertUser(ctx context.Context, user domain.UserInstance) error {
	userRow := UserRow{
		UUID:     user.UUID,
		Username: user.Username,
		Password: user.Password,
	}
	var operationErr error

	insertOperation := func() error {
		_, err := db.pool.Exec(
			ctx,
			"INSERT INTO client(uuid, name, password) VALUES($1, $2, $3)",
			userRow.UUID,
			userRow.Username,
			userRow.Password,
		)
		if err == nil {
			return nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			operationErr = domain.ErrUserAlreadyExists
			return nil
		}

		return err
	}

	if err := BackoffRetry(ctx, maxAttempts, baseDelay, insertOperation); err != nil {
		return err
	}

	return operationErr
}
