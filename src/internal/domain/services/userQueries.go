package services

import (
	"context"
	"fmt"
	"tickTackToe/internal/domain"
)

func (us *UserService) GetUserStatus(ctx context.Context, userUUID string) (domain.UserInstance, error) {
	user, err := us.Repo.FindUserByID(ctx, userUUID)
	if err != nil {
		return domain.UserInstance{}, fmt.Errorf("get user status: %w", err)
	}
	return user, nil
}
