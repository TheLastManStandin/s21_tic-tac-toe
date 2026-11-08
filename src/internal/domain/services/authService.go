package services

import (
	"context"
	"tickTackToe/internal/domain"
)

type AuthService struct {
	userService *UserService
	jwtProvider *JwtProvider
}

func NewAuthService(userService *UserService, jwtProvider *JwtProvider) *AuthService {
	return &AuthService{
		userService: userService,
		jwtProvider: jwtProvider,
	}
}

func (s *AuthService) Authenticate(
	ctx context.Context,
	request domain.JwtRequest,
) (domain.JwtResponse, error) {
	userUUID, valid := s.userService.ValidateUserPassword(ctx, request.Login, request.Password)
	if !valid {
		return domain.JwtResponse{}, domain.ErrInvalidLoginPassword
	}

	acessToken, err := s.jwtProvider.GenerateAccessToken(userUUID)
	if err != nil {
		return domain.JwtResponse{}, err
	}
	refreshToken, err := s.jwtProvider.GenerateRefreshToken(userUUID)
	if err != nil {
		return domain.JwtResponse{}, err
	}

	return domain.JwtResponse{
		Type:         "Bearer",
		AccessToken:  acessToken,
		RefreshToken: refreshToken,
	}, nil
}
