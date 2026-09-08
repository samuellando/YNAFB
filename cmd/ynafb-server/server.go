package main

import (
	"log"
	"net/http"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db"
	"samuellando.com/YNAFB/internal/http/handler"
	"samuellando.com/YNAFB/internal/http/middleware"
)

func main() {
	m := http.NewServeMux()
	db, err := dbutil.Open("./ynafb.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	queries := data.New(db)
	loginHandler := handler.CreateLoginHandler(queries)
	// Authenticated handlers
	budgetHandler := middleware.Authenticator(handler.CreateBudgetHandler(queries))
	accountHandler := middleware.Authenticator(handler.CreateAccountHandler(queries, db))
	// Login endpoints
	m.Handle("/auth/", http.StripPrefix("/auth", loginHandler))
	// API endpoints
	// Budget Routes
	m.Handle("/api/v1/budget", http.StripPrefix("/api/v1", budgetHandler))
	m.Handle("/api/v1/budget/{budget}", http.StripPrefix("/api/v1", budgetHandler))
	m.Handle("/api/v1/budget/{budget}/{month}", http.StripPrefix("/api/v1", budgetHandler))
	// Account Routes
	m.Handle("/api/v1/budget/{budget}/account", http.StripPrefix("/api/v1", accountHandler))
	m.Handle("/api/v1/budget/{budget}/account/{account}", http.StripPrefix("/api/v1", accountHandler))
	m.Handle("/api/v1/budget/{budget}/account/{account}/import", http.StripPrefix("/api/v1", accountHandler))

	log.Fatal(http.ListenAndServe(":8080", middleware.Logger(m)))
}
