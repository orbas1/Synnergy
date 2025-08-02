package server

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// RequestLogger writes basic request info using structured logging.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Info("incoming request")
		next.ServeHTTP(w, r)
	})
}

// JSONHeaders sets Content-Type application/json for all responses.
func JSONHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
