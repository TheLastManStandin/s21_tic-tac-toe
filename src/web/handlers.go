package web

import (
	"datasource"
	"domain"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func gamePostHandler(db datasource.Database, uuidRegexp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	UUID := r.PathValue("UUID")

	var req GameRequest
	var resp GameResponse

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateRequest(req, UUID, *uuidRegexp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := domain.HandleTurn(req.toDomain(), db, UUID)
	if err != nil {
		http.Error(w, "Error handling turn: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp.UUID = ret.UUID
	resp.Table = ret.Table

	resp.Result = gameResult(int(ret.Status))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}

func registrationHandler(db datasource.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="api"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}

		bcryptPass, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		user := datasource.User{
			Username:     username,
			UUID:         uuid.New().String(),
			HashPassword: bcryptPass,
		}

		err := db.AddUser(user)
		if err != nil {
			http.Error(w, "Error adding user: "+err.Error(), http.StatusBadRequest)
		}
		w.Write([]byte("Registration successful"))
	}
}
