package middleware

import (
	"fakereddit/redditclone/pkg/handlers"
	"log"
	"net/http"
)

func Panic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("ERROR: panicMiddleware: recovered", err)
				handlers.JSONErrorBuilder(w, "Internal server error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
