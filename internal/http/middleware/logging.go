package middleware

import (
	"log"
	"net/http"
	"time"

	"samuellando.com/YNAFB/internal/db/querycount"
)

func Logger(h http.Handler) http.Handler {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx, counter := querycount.NewContext(r.Context())
		r = r.WithContext(ctx)
		// call the original http.Handler we're wrapping
		h.ServeHTTP(w, r)
		log.Printf("[%s] %s queries=%d queryTime=%s duration=%s", r.Method, r.URL.String(), counter.Value(), counter.QueryTime(), time.Since(start))
	})
	return fn
}
