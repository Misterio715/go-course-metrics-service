package handler

import (
	"log"
	"net/http"
)

func LogRequests(logger *log.Logger, next http.Handler) http.Handler {
	if logger == nil {
		logger = log.Default()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("incoming request method=%s path=%s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
