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
	db, err := dbutil.Open("./ynafb.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	queries := data.New(db)
	budgetHandler := handler.Budget{Queires: queries}
	loginHandler := handler.Login{Queires: queries}
	accountHandler := handler.Account{Queries: queries, DB: db}
	http.HandleFunc("POST /signup", middleware.Logger(loginHandler.CreateLogin)) 
	http.HandleFunc("POST /authenticate", middleware.Logger(loginHandler.Authenticate)) 
	http.HandleFunc("POST /deauthenticate", middleware.Logger(loginHandler.Deauthenticate)) 
	// Budget endpoints
	http.HandleFunc("POST /budget", middleware.Logger(middleware.Authenticator(budgetHandler.CreateBudget))) 
	http.HandleFunc("GET /budget", middleware.Logger(middleware.Authenticator(budgetHandler.ListBudgets))) 
	http.HandleFunc("PUT /budget/{id}", middleware.Logger(middleware.Authenticator(budgetHandler.UpdateBudget))) 
	http.HandleFunc("GET /budget/{id}", middleware.Logger(middleware.Authenticator(budgetHandler.GetBudgetMonth))) 
	http.HandleFunc("GET /budget/{id}/{month}", middleware.Logger(middleware.Authenticator(budgetHandler.GetBudgetMonth))) 
	http.HandleFunc("DELETE /budget/{id}", middleware.Logger(middleware.Authenticator(budgetHandler.DeleteBudget)) )
	// Account endpoints
	http.HandleFunc("GET /budget/{budget}/account", middleware.Logger(middleware.Authenticator(accountHandler.ListAccounts))) 
	http.HandleFunc("POST /budget/{budget}/account", middleware.Logger(middleware.Authenticator(accountHandler.CreateAccount))) 
	http.HandleFunc("POST /budget/{budget}/account/{id}/import", middleware.Logger(middleware.Authenticator(accountHandler.Import))) 
	http.HandleFunc("GET /budget/{budget}/account/{id}", middleware.Logger(middleware.Authenticator(accountHandler.GetAccount)) )
	http.HandleFunc("PUT /budget/{budget}/account/{id}", middleware.Logger(middleware.Authenticator(accountHandler.UpdateAccount))) 
	http.HandleFunc("DELETE /budget/{budget}/account/{id}", middleware.Logger(middleware.Authenticator(accountHandler.DeleteAccount))) 
	log.Fatal(http.ListenAndServe(":8080", nil))
}
