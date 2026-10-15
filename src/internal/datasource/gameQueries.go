package datasource

import (
	"context"
	"errors"
	"tickTackToe/internal/domain"
)

func (db *PostgresDatabase) ListAvailableGames(ctx context.Context, playerUUID string) ([]domain.GameInstance, error) {
	var games []domain.GameInstance
	operation := func() error {
		games = make([]domain.GameInstance, 0)
		rows, err := db.pool.Query(ctx, `
			SELECT uuid, gametable, gamestate, minimax, firstplayer,
			       COALESCE(secondplayer::text, '')
			FROM games
			WHERE minimax = FALSE AND gamestate = $1
			  AND secondplayer IS NULL AND firstplayer IS NOT NULL
			  AND firstplayer <> $2
			ORDER BY uuid`, domain.WaitingForPlayers, playerUUID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var game domain.GameInstance
			var table [][]int32
			if err := rows.Scan(&game.UUID, &table, &game.State, &game.Minimax,
				&game.FirstPlayerUUID, &game.SecondPlayerUUID); err != nil {
				return err
			}
			if len(table) != 3 {
				return errors.New("invalid table dimensions in database")
			}
			for x, row := range table {
				if len(row) != 3 {
					return errors.New("invalid table dimensions in database")
				}
				for y, value := range row {
					game.Table[x][y] = int(value)
				}
			}
			games = append(games, game)
		}
		return rows.Err()
	}
	if err := BackoffRetry(ctx, maxAttempts, baseDelay, operation); err != nil {
		return nil, err
	}
	return games, nil
}
