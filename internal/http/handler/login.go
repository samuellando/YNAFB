package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"samuellando.com/YNAFB/internal/data"
	"samuellando.com/YNAFB/internal/auth"
)

type Login struct {
	*http.ServeMux
	queries *data.Queries
}

func CreateLoginHandler(queries *data.Queries) http.Handler {
	l := Login{
		queries: queries,
		ServeMux: http.NewServeMux(),
	}
	l.HandleFunc("POST /signup", l.CreateLogin) 
	l.HandleFunc("POST /authenticate", l.Authenticate) 
	l.HandleFunc("POST /deauthenticate", l.Deauthenticate) 
	return l
}


func (b Login) CreateLogin(w http.ResponseWriter, req *http.Request) {
	log.Println("Create login")
	// Parse the input
	params := data.CreateLoginParams{}
	data, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(data, &params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(params.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}
	params.Password, err = auth.HashPassword(params.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Query the db
	_, err = b.queries.CreateLogin(req.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (b Login) Authenticate(w http.ResponseWriter, req *http.Request) {
	log.Println("authentication")
	// Parse the input
	input := data.CreateLoginParams{}
	d, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(d, &input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Query the db
	login, err := b.queries.GetLoginByUsername(req.Context(), data.GetLoginByUsernameParams{
		Username: input.Username,
	})
	if err != nil {
		//dummy
		auth.CheckPassword("dummy", "dummy")
		http.Error(w, "Denied", http.StatusUnauthorized)
		return
	}
	// Check the password hash
	if !auth.CheckPassword(input.Password, login.Password) {
		http.Error(w, "Denied", http.StatusUnauthorized)
		return
	}
	// Send a JWT token
	token, err := auth.GenerateJWT(strconv.Itoa(int(login.ID)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	auth.SetJWTCookie(w, token)
}

func (b Login) Deauthenticate(w http.ResponseWriter, req *http.Request) {
	log.Println("deauthentication")
	cookie, err := auth.GetJWTCookie(req)
	if err != nil {
		return
	}
	auth.DevalidateJWT(cookie.Value)
	auth.UnsetJWTCookie(w)
}
