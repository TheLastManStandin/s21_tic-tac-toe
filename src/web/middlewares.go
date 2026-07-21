package web

import (
	"datasource"
	"net/http"
	"regexp"
)

func authenticationMiddleware(
	db datasource.Database,
	uuidRegexp *regexp.Regexp,
	wrappedFunc func(db datasource.Database, uuidRegexp *regexp.Regexp, w http.ResponseWriter, r *http.Request),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
