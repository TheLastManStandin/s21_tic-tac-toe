package handlers

import (
	"net/http"
	"strings"
	"tickTackToe/internal/domain/services"
	"tickTackToe/internal/web/httpresponse"

	"go.uber.org/zap"
)

type AuthHandler struct {
	service *services.UserService
	logger  *zap.Logger
}

func NewAuthHandler(service *services.UserService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		logger:  logger,
	}
}

func (h AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		httpresponse.WriteUnauthorized(w)
		return
	}
	uuid, err := h.service.RegisterUser(r.Context(), username, password)

	if err != nil {
		h.logger.Error("Failed to register user", zap.Error(err))
		httpresponse.WriteError(w, err)
		return
	}
	_, _ = w.Write([]byte(uuid))
}

func (h AuthHandler) Auth(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok || strings.TrimSpace(username) == "" || password == "" {
		httpresponse.WriteUnauthorized(w)
		return
	}

	userUUID, authenticated := h.service.AuthenticateUser(
		r.Context(),
		username,
		password,
	)
	if !authenticated {
		httpresponse.WriteUnauthorized(w)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(userUUID))
}
