package services

import (
	"context"
	"fmt"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/repositorys"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Repo   repositorys.UserRepository
	logger *zap.Logger
}

func NewUserService(repo repositorys.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{
		Repo:   repo,
		logger: logger,
	}
}

func (us *UserService) AuthenticateUser(ctx context.Context, username string, password string) (string, bool) {
	user, err := us.Repo.GetUser(ctx, username)
	if err != nil {
		return "", false
	}

	if cmperr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); cmperr != nil {
		return "", false
	}

	return user.UUID, true
}

func (us *UserService) RegisterUser(ctx context.Context, username string, password string) (string, error) {
	bcryptPass, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := domain.UserInstance{
		Username: username,
		Password: string(bcryptPass),
		UUID:     uuid.New().String(),
	}

	if err := us.Repo.InsertUser(ctx, user); err != nil {
		return "", err
	}
	return user.UUID, nil
}

func (us *UserService) GetUserStatus(ctx context.Context, userUUID string) (domain.UserInstance, error) {
	user, err := us.Repo.FindUserByID(ctx, userUUID)
	if err != nil {
		return domain.UserInstance{}, fmt.Errorf("get user status: %w", err)
	}
	return user, nil
}
