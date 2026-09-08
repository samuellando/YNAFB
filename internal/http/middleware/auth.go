package middleware

import (
	"context"
	"strconv"
	"net/http"

	"samuellando.com/YNAFB/internal/auth"
)

func Authenticator(h http.Handler) http.Handler {
	fn := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// call the original http.Handler we're wrapping
		cookie, err := auth.GetJWTCookie(req)
		if err != nil {
			http.Error(w, "Denied", http.StatusUnauthorized)
			return
		}
		subject, newToken, err := auth.ParseJWT(cookie.Value)
		if err != nil {
			http.Error(w, "Denied", http.StatusUnauthorized)
			return
		}
		loginID, err := strconv.Atoi(subject)
		if err != nil {
			http.Error(w, "Denied", http.StatusInternalServerError)
			return
		}
		// Update the context
		ctx := context.WithValue(req.Context(), "loginID", int64(loginID))
		req = req.WithContext(ctx)
		// Set the new token
		auth.SetJWTCookie(w, newToken)
		// Call the handler
		h.ServeHTTP(w, req)
	})
	return fn
}
