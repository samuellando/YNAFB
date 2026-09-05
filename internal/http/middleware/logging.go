package middleware

import (
	"log"
	"net/http"
)

func Logger(h http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s", r.Method, r.URL.String()) 
		// call the original http.Handler we're wrapping
		h(w, r)
	}
	return fn
}
