package handlers

import (
	"net/http"
	"strings"
	"tickTackToe/internal/domain/services"

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
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	uuid, err := h.service.RegisterUser(r.Context(), username, password)

	if err != nil {
		http.Error(w, "Failed to register user: "+err.Error(), http.StatusInternalServerError)
		h.logger.Error("Failed to register user", zap.Error(err))
		return
	}
	_, _ = w.Write([]byte(uuid))
}

func (h AuthHandler) Auth(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok || strings.TrimSpace(username) == "" || password == "" {
		writeUnauthorized(w)
		return
	}

	userUUID, authenticated := h.service.AuthenticateUser(
		r.Context(),
		username,
		password,
	)
	if !authenticated {
		writeUnauthorized(w)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(userUUID))
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set(
		"WWW-Authenticate",
		`Basic realm="tic-tac-toe", charset="UTF-8"`,
	)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}
