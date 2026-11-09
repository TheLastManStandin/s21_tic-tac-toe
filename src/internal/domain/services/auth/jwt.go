package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/services"
)

type JwtAuthService struct {
	userService *services.UserService
	jwtProvider *domain.JwtProvider
}

func NewJwtAuthService(userService *services.UserService, jwtProvider *domain.JwtProvider) *JwtAuthService {
	return &JwtAuthService{
		userService: userService,
		jwtProvider: jwtProvider,
	}
}

func (s *JwtAuthService) Authenticate(r *http.Request) (any, error) {
	var request domain.JwtRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		return domain.JwtResponse{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return domain.JwtResponse{}, domain.ErrInvalidLoginPassword
	}

	if strings.TrimSpace(request.Login) == "" || request.Password == "" {
		return domain.JwtResponse{}, domain.ErrEmptyLoginPassword
	}

	userUUID, valid := s.userService.ValidateUserPassword(r.Context(), request.Login, request.Password)
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
