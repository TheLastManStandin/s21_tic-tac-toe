package api

import "net/http"

func WebInit() {
	mux := http.NewServeMux()

	mux.HandleFunc("/hc", heathCheck)
	//mux.HandleFunc("/game")

	_ = http.ListenAndServe("localhost:5005", mux)
}

func heathCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK\n"))
}
