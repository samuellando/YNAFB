package auth

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var key []byte
var m sync.Mutex

var validTokens map[string]time.Time

func init() {
	m.Lock()
	defer m.Unlock()
	key = make([]byte, 32)
	rand.Read(key)

	validTokens = make(map[string]time.Time)

	// Clean up the token cache every mintue
	go func() {
		ticker := time.NewTicker(time.Minute)
		for {
			<-ticker.C
			clearExpiredTokens()
		}
	}()
}

func clearExpiredTokens() {
	m.Lock()
	defer m.Unlock()
	log.Print("[AUTH] Cleaning token cache...")
	newValidTokens := make(map[string]time.Time)
	for token, expires := range validTokens {
		if time.Now().Before(expires) {
			newValidTokens[token] = expires
		}
	}
	validTokens = newValidTokens
}

func GenerateJWT(subject string) (string, error) {
	m.Lock()
	defer m.Unlock()
	now := time.Now()
	expires := time.Now().Add(time.Hour)
	claims := &jwt.RegisteredClaims{
		Subject:   subject,
		ID:        uuid.NewString(),
		IssuedAt:  &jwt.NumericDate{Time: now},
		ExpiresAt: &jwt.NumericDate{Time: expires},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(key)
	if err != nil {
		return "", err
	}
	// Store the token locally
	validTokens[s] = expires
	return s, nil
}

func valid(token string) bool {
	m.Lock()
	defer m.Unlock()
	// Check if the token is stored, or is expired
	if expires, ok := validTokens[token]; !ok || time.Now().After(expires) {
		return false
	}
	return true
}

func ParseJWT(token string) (subject string, newToken string, err error) {
	if !valid(token) {
		return "", "", fmt.Errorf("Denied")
	}
	// Parse the subject
	claims := &jwt.RegisteredClaims{}
	_, err = jwt.ParseWithClaims(token, claims, func(*jwt.Token) (any, error) {
		return key, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return "", "", fmt.Errorf("Denied")
	}
	// Generate a new JWT and return
	newToken, err = GenerateJWT(claims.Subject)
	if err != nil {
		return "", "", err
	}
	return claims.Subject, newToken, nil
}

func DevalidateJWT(token string) {
	m.Lock()
	defer m.Unlock()
	delete(validTokens, token)
}

func GetJWTCookie(req *http.Request) (*http.Cookie, error) {
	return req.Cookie("jwt")
}

func SetJWTCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Secure: true,
		HttpOnly: true,
		Path: "/",
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

func UnsetJWTCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    "",
		MaxAge: -1,
		Path: "/",
	}
	http.SetCookie(w, cookie)
}
