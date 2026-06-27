package api

import (
	"context"
	"domain"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"go.uber.org/fx"
)

//func WebInit() {
//	mux := http.NewServeMux()
//
//	mux.HandleFunc("/hc", heathCheck)
//	mux.HandleFunc("/game/{UUID}", handleGamePost)
//
//	srv := &http.Server{
//		Addr:    ":5005",
//		Handler: mux,
//	}
//	fmt.Println("WELCOME TO THE GAME, TRAVELER")
//	defer srv.Shutdown(context.Background())
//	srv.ListenAndServe()
//	//_ = http.ListenAndServe("localhost:5005", mux)
//}

func InitHTTPServer(lc fx.Lifecycle) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/hc", heathCheck)
	mux.HandleFunc("/game/{UUID}", handleGamePost)

	srv := &http.Server{
		Addr:    ":5005",
		Handler: mux,
	}
	//fmt.Println("WELCOME TO THE GAME, TRAVELER")

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			fmt.Println("WELCOME TO THE GAME, TRAVELER. TRY " + srv.Addr)
			go func() {
				if err := srv.ListenAndServe(); err != http.ErrServerClosed {
					log.Fatalf("ListenAndServe(): %s", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("SHUTDOWN")
			return srv.Shutdown(ctx)
		},
	})

	return srv
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
