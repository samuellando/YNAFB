package middleware

import (
	"log"
	"net/http"
)

func Logger(h http.Handler) http.Handler {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s", r.Method, r.URL.String()) 
		// call the original http.Handler we're wrapping
		h.ServeHTTP(w, r)
	})
	return fn
}
