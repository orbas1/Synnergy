package middleware

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logrus.Infof("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}
