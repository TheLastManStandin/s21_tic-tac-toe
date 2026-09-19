package repositorys

import (
	"context"
	"tickTackToe/internal/domain"
)

type UserRepository interface {
	GetUser(ctx context.Context, username string) (domain.UserInstance, error)
	InsertUser(ctx context.Context, user domain.UserInstance) error
}
