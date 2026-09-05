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
	http.HandleFunc("POST /login", middleware.Logger(loginHandler.CreateLogin)) 
	http.HandleFunc("POST /budget", middleware.Logger(budgetHandler.CreateBudget)) 
	http.HandleFunc("GET /budget", middleware.Logger(budgetHandler.ListBudgets)) 
	http.HandleFunc("PUT /budget/{id}", middleware.Logger(budgetHandler.UpdateBudget)) 
	http.HandleFunc("GET /budget/{id}", middleware.Logger(budgetHandler.GetBudgetMonth)) 
	http.HandleFunc("GET /budget/{id}/{month}", middleware.Logger(budgetHandler.GetBudgetMonth)) 
	http.HandleFunc("DELETE /budget/{id}", middleware.Logger(budgetHandler.DeleteBudget)) 
	http.HandleFunc("GET /budget/{budget}/account", middleware.Logger(accountHandler.ListAccounts)) 
	http.HandleFunc("POST /budget/{budget}/account", middleware.Logger(accountHandler.CreateAccount)) 
	http.HandleFunc("POST /budget/{budget}/account/{id}/import", middleware.Logger(accountHandler.Import)) 
	http.HandleFunc("GET /budget/{budget}/account/{id}", middleware.Logger(accountHandler.GetAccount)) 
	http.HandleFunc("PUT /budget/{budget}/account/{id}", middleware.Logger(accountHandler.UpdateAccount)) 
	http.HandleFunc("DELETE /budget/{budget}/account/{id}", middleware.Logger(accountHandler.DeleteAccount)) 
	log.Fatal(http.ListenAndServe(":8080", nil))
}
