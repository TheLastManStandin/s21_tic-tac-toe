package web

import (
	"encoding/json"
	"net/http"
)

func (h GamePostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req GameRequest
	var resp GameResponse

	user, authErr := h.userService.Authentication(*r)
	_ = user
	if authErr != nil {
		w.Header().Set("WWW-Authenticate", `Basic realm="api"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	UUID := r.PathValue("UUID")

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.gameService.GetUpdatedGame(UUID, req.toDomain())
	if err != nil {
		http.Error(w, "Error handling turn: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp.UUID = ret.UUID
	resp.Table = ret.Table
	resp.Result = gameResult(ret.State)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}
