package httpresponse

import (
	"errors"
	"net/http"
	"tickTackToe/internal/domain"
)

func WriteError(w http.ResponseWriter, err error) {
	status, message := ErrorResponse(err)
	http.Error(w, message, status)
}

func ErrorResponse(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrGameAccessDenied):
		return http.StatusForbidden, domain.ErrGameAccessDenied.Error()
	case errors.Is(err, domain.ErrGameNotFound):
		return http.StatusNotFound, domain.ErrGameNotFound.Error()
	}

	for _, conflict := range []error{
		domain.ErrUserAlreadyExists,
		domain.ErrCannotConnectToGame,
		domain.ErrCantConnectGameWithRobot,
		domain.ErrGameAlreadyStarted,
		domain.ErrAlreadyInGame,
		domain.ErrNotYourTurn,
		domain.ErrWaitingForPlayers,
		domain.ErrCellOccupied,
	} {
		if errors.Is(err, conflict) {
			return http.StatusConflict, conflict.Error()
		}
	}

	return http.StatusInternalServerError, "internal server error"
}

func WriteUnauthorized(w http.ResponseWriter) {
	w.Header().Set(
		"WWW-Authenticate",
		`Basic realm="tic-tac-toe", charset="UTF-8"`,
	)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}
