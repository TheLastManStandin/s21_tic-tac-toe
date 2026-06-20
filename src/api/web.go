package api

import (
	"domain"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

func WebInit() {
	mux := http.NewServeMux()

	mux.HandleFunc("/hc", heathCheck)
	mux.HandleFunc("/game/{UUID}", handleGamePost)

	fmt.Println("WELCOME TO THE GAME, TRAVELER")
	_ = http.ListenAndServe("localhost:5005", mux)
}

func heathCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK\n"))
}

func handleGamePost(w http.ResponseWriter, r *http.Request) {
	UUID := r.PathValue("UUID")

	var req GameRequest
	if err := validateUUID(UUID, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err, ret := domain.HandleTurn(req.ToDomainGame())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(ret + "\n"))
}

func validateUUID(UUID string, req *GameRequest) error {
	var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(UUID) {
		return errors.New("invalid UUID format")
	}
	req.ID = UUID
	return nil
}
