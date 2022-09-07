package middleware

import (
	"log"
	"net/http"
)

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		log.Printf("New request: method:[%v] remote_addr [%v] url [%v]", r.Method, r.RemoteAddr, r.URL.Path)
	})
}
