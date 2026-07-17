package web

import (
	"datasource"
	"domain"
	"net/http"
	"regexp"
)

func authenticationMiddleware(
	db datasource.Database,
	uuidRegexp *regexp.Regexp,
	wrappedFunc func(db datasource.Database, uuidRegexp *regexp.Regexp, w http.ResponseWriter, r *http.Request),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok && domain.VerifyUserPass(username, password, db) {
			wrappedFunc(db, uuidRegexp, w, r)
		} else {
			w.Header().Set("WWW-Authenticate", `Basic realm="api"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	}
}
