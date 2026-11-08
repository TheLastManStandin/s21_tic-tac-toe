package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"tickTackToe/internal/domain"
	"tickTackToe/internal/domain/services"
	"tickTackToe/internal/web/httpresponse"

	"go.uber.org/zap"
)

type AuthHandler struct {
	userService *services.UserService
	authService *services.AuthService
	logger      *zap.Logger
}

func NewAuthHandler(
	userService *services.UserService,
	logger *zap.Logger,
	authService *services.AuthService,
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
		h.logger.Error("Failed to register user", zap.Error(err))
		httpresponse.WriteError(w, err)
		return
	}
	_, _ = w.Write([]byte(userUUID))
}

func (h AuthHandler) Auth(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request domain.JwtRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, "expected one JSON object", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(request.Login) == "" || request.Password == "" {
		http.Error(w, "login and password are required", http.StatusBadRequest)
		return
	}

	response, err := h.authService.Authenticate(r.Context(), request)
	if errors.Is(err, domain.ErrInvalidLoginPassword) {
		httpresponse.WriteUnauthorized(w)
		return
	}
	if err != nil {
		h.logger.Error("authentication failed", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to write auth response", zap.Error(err))
	}
}
