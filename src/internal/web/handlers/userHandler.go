package handlers

import (
	"encoding/json"
	"net/http"
	"tickTackToe/internal/domain/services"
	"tickTackToe/internal/web/httpresponse"
	"tickTackToe/internal/web/requestcontext"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserHandler struct {
	service *services.UserService
	logger  *zap.Logger
}

func NewUserHandler(service *services.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{service: service, logger: logger}
}

func (h *UserHandler) GetUserStatus(w http.ResponseWriter, r *http.Request) {
	currentUserUUID, ok := requestcontext.UserUUID(r.Context())
	if !ok || currentUserUUID == "" {
		httpresponse.WriteUnauthorized(w)
		return
	}
	parsedUUID, err := uuid.Parse(r.PathValue("UUID"))
	if err != nil {
		http.Error(w, "invalid user UUID", http.StatusBadRequest)
		return
	}
	user, err := h.service.GetUserStatus(r.Context(), parsedUUID.String())
	if err != nil {
		h.logger.Info("failed to get user status", zap.Error(err))
		httpresponse.WriteError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(UserResponse{UUID: user.UUID, Username: user.Username}); err != nil {
		h.logger.Error("failed to write user status", zap.Error(err))
	}
}
