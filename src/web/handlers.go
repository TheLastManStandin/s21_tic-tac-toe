package web

import (
	"datasource"
	"domain"
	"encoding/json"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
)

type GamePostHandler struct {
	log         *zap.Logger
	gameService domain.GameService
}

func gameResult(status datasource.GameState) string {
	//switch status {
	//case datasource.Draw:
	//	return "draw"
	//case datasource.:
	//	return "player_won"
	//case domain.MinimaxWin:
	//	return "computer_won"
	//default:
	//	return "in_progress"
	//}
	return "Сделай статусы"
}

func authHandler(db datasource.Database, uuidRegexp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	username, _, _ := r.BasicAuth()

	user, err := db.GetUserByName(username)

	if err != nil {
		http.Error(w, "Error getting user: "+err.Error(), http.StatusBadRequest)
	}
	w.Write([]byte(user.UUID + "\n"))
}

func registrationHandler(db datasource.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		//username, password, ok := r.BasicAuth()
		req := SignUpRequest{}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			w.Header().Set("WWW-Authenticate", `Basic realm="api"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if len(req.Username) < 4 {
			w.Write([]byte("Username must be at least 4 characters"))
		}
		if len(req.Password) < 4 {
			w.Write([]byte("Password must be at least 4 characters"))
		}

		bcryptPass, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

		user := datasource.User{
			Username:     req.Username,
			UUID:         uuid.New().String(),
			HashPassword: bcryptPass,
		}

		err := db.AddUser(user)
		if err != nil {
			http.Error(w, "Error adding user: "+err.Error(), http.StatusBadRequest)
		} else {
			w.Write([]byte("Registration successful\n"))
		}
	}
}
