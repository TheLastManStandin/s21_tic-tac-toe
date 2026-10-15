package datasource

import (
	"context"
	"errors"
	"tickTackToe/internal/domain"

	"github.com/jackc/pgx/v5"
)

func (db *PostgresDatabase) FindUserByID(ctx context.Context, userUUID string) (domain.UserInstance, error) {
	var user domain.UserInstance
	var queryErr error
	operation := func() error {
		user = domain.UserInstance{}
		queryErr = db.pool.QueryRow(ctx,
			"SELECT uuid, name FROM client WHERE uuid = $1", userUUID,
		).Scan(&user.UUID, &user.Username)
		if errors.Is(queryErr, pgx.ErrNoRows) {
			return nil
		}
		return queryErr
	}
	if err := BackoffRetry(ctx, maxAttempts, baseDelay, operation); err != nil {
		return domain.UserInstance{}, err
	}
	if errors.Is(queryErr, pgx.ErrNoRows) {
		return domain.UserInstance{}, domain.ErrUserNotFound
	}
	return user, nil
}
