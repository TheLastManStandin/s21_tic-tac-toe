package auth

import (
	"net/http"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/services"
)

type BasicAuthService struct {
	userService *services.UserService
}

func NewBasicAuthService(userService *services.UserService) *BasicAuthService {
	return &BasicAuthService{
		userService: userService,
	}
}

func (s *BasicAuthService) Login(r *http.Request) (any, error) {
	login, password, ok := r.BasicAuth()
	if !ok || login == "" || password == "" {
		return "", domain.ErrEmptyLoginPassword
	}
	userUUID, valid := s.userService.ValidateUserPassword(r.Context(), login, password)
	if !valid {
		return "", domain.ErrInvalidLoginPassword
	}
	return userUUID, nil
}

func (s *BasicAuthService) Authorize(r *http.Request) (any, error) {
	login, password, ok := r.BasicAuth()
	if !ok || login == "" || password == "" {
		return "", domain.ErrEmptyLoginPassword
	}
	userUUID, valid := s.userService.ValidateUserPassword(r.Context(), login, password)
	if !valid {
		return "", domain.ErrInvalidLoginPassword
	}
	return userUUID, nil
}
