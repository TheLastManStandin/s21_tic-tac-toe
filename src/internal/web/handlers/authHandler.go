package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/services"
	"tickTackToe/internal/web/utils/httpresponse"
	"tickTackToe/internal/web/utils/requestcontext"

	"go.uber.org/zap"
)

type AuthHandler struct {
	userService *services.UserService
	authService domain.AuthService
	logger      *zap.Logger
}

func NewAuthHandler(
	userService *services.UserService,
	logger *zap.Logger,
	authService domain.AuthService,
) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		authService: authService,
		logger:      logger,
	}
}

func (h AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		httpresponse.WriteUnauthorized(w)
		return
	}
	userUUID, err := h.userService.RegisterUser(r.Context(), username, password)

	if err != nil {
		if !errors.Is(err, domain.ErrUserAlreadyExists) {
			h.logger.Error("Failed to register user", zap.Error(err))
		}
		httpresponse.WriteError(w, err)
		return
	}
	_, _ = w.Write([]byte(userUUID))
}

func (h AuthHandler) Auth(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	response, err := h.authService.Login(r)
	if errors.Is(err, domain.ErrInvalidLoginPassword) || errors.Is(err, domain.ErrEmptyLoginPassword) {
		httpresponse.WriteUnauthorized(w)
		return
	}
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		h.logger.Error("authentication failed", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to write auth response", zap.Error(err))
	}
}

func (h AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	response, err := h.authService.Refresh(r)
	if errors.Is(err, domain.ErrInvalidRefreshToken) {
		httpresponse.WriteUnauthorized(w)
		return
	}
	if err != nil {
		h.logger.Error("refresh failed", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to write refresh response", zap.Error(err))
	}
}

func (h AuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := requestcontext.UserUUID(r.Context())
	if !ok || userUUID == "" {
		h.logger.Error("failed to get user UUID from request")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(userUUID); err != nil {
		h.logger.Error("failed to write user info response", zap.Error(err))
	}
}
