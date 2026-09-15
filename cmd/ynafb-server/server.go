package main

import (
	"log"
	"net/http"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db"
	"samuellando.com/YNAFB/internal/http/handler"
	"samuellando.com/YNAFB/internal/http/middleware"
	"samuellando.com/YNAFB/internal/http/api"
	"samuellando.com/YNAFB/internal/config"
)

func main() {
	err := config.LoadConfigFromFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	m := http.NewServeMux()
	db, err := dbutil.Open("./ynafb.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	queries := data.New(db)


	loginHandler := handler.CreateLoginHandler(queries)
	// // Authenticated handlers
	m.Handle("/auth/", http.StripPrefix("/auth", loginHandler))
	// API endpoints
	server := api.NewServer(db)
	h := middleware.Authenticator(api.Handler(api.NewStrictHandler(server, nil)))
	m.Handle("/api/v1/", http.StripPrefix("/api/v1", h))
	// SPA (embedded frontend); more specific patterns above take precedence
	m.Handle("/", spaHandler())

	log.Fatal(http.ListenAndServe(":8080", middleware.Logger(m)))
}
