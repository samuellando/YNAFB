package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db/types"
)

type Budget struct {
	*http.ServeMux
	queries *data.Queries
}

func CreateBudgetHandler(queries *data.Queries) http.Handler {
	b := Budget{
		queries: queries,
		ServeMux: http.NewServeMux(),
	}
	b.HandleFunc("POST /budget", b.CreateBudget) 
	b.HandleFunc("GET /budget", b.ListBudgets) 
	b.HandleFunc("PUT /budget/{id}", b.UpdateBudget) 
	b.HandleFunc("GET /budget/{id}", b.GetBudgetMonth) 
	b.HandleFunc("GET /budget/{id}/{month}", b.GetBudgetMonth) 
	b.HandleFunc("DELETE /budget/{id}", b.DeleteBudget)
	return b
}

type GetBudgetMonthResponse struct {
	Month time.Time
	Summary    data.GetBudgetMonthSummaryRow
	Categories []data.ListBudgetMonthCategoriesRow
	Goals      []data.ListGoalsValuesRow
}

func (b Budget) CreateBudget(w http.ResponseWriter, req *http.Request) {
	log.Println("Create budget")
	// Parse the input
	params := data.CreateBudgetParams{}
	data, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(data, &params)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Use the loginID from the context
	params.LoginID, err = getLoginID(req) 
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Query the db
	budget, err := b.queries.CreateBudget(req.Context(), params)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Send response
	res, err := json.Marshal(budget)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func (b Budget) UpdateBudget(w http.ResponseWriter, req *http.Request) {
	log.Println("Update budget")
	// Parse the input
	params := data.UpdateBudgetParams{}
	data, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(data, &params)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Use the id from the URL
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	params.ID = int64(id)
	// Use the loginID from the context
	params.LoginID, err = getLoginID(req) 
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Query the db
	_, err = b.queries.UpdateBudget(req.Context(), params)
	if errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		http.Error(w, "Nothing to update", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (b Budget) DeleteBudget(w http.ResponseWriter, req *http.Request) {
	log.Println("Delete budget")
	// Parse the input
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	params := data.DeleteBudgetParams{ID: int64(id)}
	// Use the loginID from the context
	params.LoginID, err = getLoginID(req) 
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Query the db
	err = b.queries.DeleteBudget(req.Context(), params)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (b Budget) ListBudgets(w http.ResponseWriter, req *http.Request) {
	log.Println("List budgets")
	// Parse the input
	// Use the loginID from the context
	id, err := getLoginID(req) 
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := data.ListBudgetsParams{
		LoginID: id,
	}
	budgets, err := b.queries.ListBudgets(req.Context(), params)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Send response
	res, err := json.Marshal(budgets)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func (b Budget) GetBudgetMonth(w http.ResponseWriter, req *http.Request) {
	log.Println("Get budget month")
	// Parse the input
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	month := req.PathValue("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	monthTime, err := time.Parse("2006-01", month)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	loginID, err := getLoginID(req) 
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := data.GetBudgetMonthSummaryParams{
		ID: int64(id),
		Month:    types.UnixTime{Time: monthTime},
		LoginID: loginID,
	}
	sumamry, err := b.queries.GetBudgetMonthSummary(req.Context(), params)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	catParams := data.ListBudgetMonthCategoriesParams{
		ID: int64(id),
		Month:    types.UnixTime{Time: monthTime},
		LoginID: loginID,
	}
	categories, err := b.queries.ListBudgetMonthCategories(req.Context(), catParams)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	goalsParams := data.ListGoalsValuesParams{
		ID: int64(id),
		Month:    types.UnixTime{Time: monthTime},
		LoginID: loginID,
	}
	goals, err := b.queries.ListGoalsValues(req.Context(), goalsParams)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Send response
	res, err := json.Marshal(GetBudgetMonthResponse{
		Month: monthTime,
		Summary:    sumamry,
		Categories: categories,
		Goals:      goals,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}
