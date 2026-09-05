package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db/types"
	"samuellando.com/YNAFB/internal/importer"
)

type Account struct {
	Queries *data.Queries
	DB      *sql.DB
}

type GetAccountResponce struct {
	Summary      data.GetAccountBalancesRow
	Transactions []data.ListAccountTransactionsRow
}

func getAccountRefsFromPath(req *http.Request) (int64, int64, error) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		return 0, 0, err
	}
	budgetID, err := strconv.Atoi(req.PathValue("budget"))
	if err != nil {
		return 0, 0, err
	}
	return int64(id), int64(budgetID), nil
}

func (b Account) CreateAccount(w http.ResponseWriter, req *http.Request) {
	log.Println("Create account")
	// Parse the input
	params := data.CreateAccountParams{}
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
	budgetID, err := strconv.Atoi(req.PathValue("budget"))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	params.BudgetID = int64(budgetID)
	// Query the db
	budget, err := b.Queries.CreateAccount(req.Context(), params)
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

func (b Account) UpdateAccount(w http.ResponseWriter, req *http.Request) {
	log.Println("Update account")
	// Parse the input
	params := data.UpdateAccountParams{}
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
	// Use the ids from the URL
	id, budgetID, err := getAccountRefsFromPath(req)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	params.ID = id
	params.BudgetID = budgetID

	// Query the db
	n, err := b.Queries.UpdateAccount(req.Context(), params)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if n == 0 {
		err := errors.New("Nothing to update")
		log.Println()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (b Account) DeleteAccount(w http.ResponseWriter, req *http.Request) {
	log.Println("Delete account")

	params := data.DeleteAccountParams{}
	// Use the ids from the URL
	id, budgetID, err := getAccountRefsFromPath(req)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	params.ID = id
	params.BudgetID = budgetID
	// Query the db
	err = b.Queries.DeleteAccount(req.Context(), params)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (b Account) ListAccounts(w http.ResponseWriter, req *http.Request) {
	log.Println("List accounts")
	// Parse the input
	params := data.ListAccountsBalancesParams{}
	budgetID, err := strconv.Atoi(req.PathValue("budget"))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	params.BudgetID = int64(budgetID)
	// Query the db
	accounts, err := b.Queries.ListAccountsBalances(req.Context(), params)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Send response
	res, err := json.Marshal(accounts)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func (b Account) GetAccount(w http.ResponseWriter, req *http.Request) {
	log.Println("Get account")
	// Use the ids from the URL
	id, budgetID, err := getAccountRefsFromPath(req)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Query the db
	summary, err := b.Queries.GetAccountBalances(req.Context(), data.GetAccountBalancesParams{
		BudgetID:  budgetID,
		AccountID: id,
	})
	if errors.Is(err, sql.ErrNoRows) {
		log.Println(err, budgetID, id)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	transactions, err := b.Queries.ListAccountTransactions(req.Context(), data.ListAccountTransactionsParams{
		BudgetID:  budgetID,
		AccountID: id,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Send response
	res, err := json.Marshal(GetAccountResponce{
		Summary:      summary,
		Transactions: transactions,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func (b Account) Import(w http.ResponseWriter, req *http.Request) {
	log.Println("Import transactions to account")
	// Use the ids from the URL
	id, budgetID, err := getAccountRefsFromPath(req)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, _, err := req.FormFile("statement")
	defer file.Close()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	content, err := io.ReadAll(file)
	stmt, err := importer.Parse(content)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tx, err := b.DB.BeginTx(req.Context(), nil)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	txQueries := b.Queries.WithTx(tx)
	for _, entry := range stmt.Entries {
		payeeName := strings.TrimSpace(entry.Payee)
		if payeeName == "" {
			payeeName = "unknown"
		}

		payeeID, err := resolveOrCreatePayeeID(req.Context(), txQueries, budgetID, payeeName)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = txQueries.CreateTrx(req.Context(), data.CreateTrxParams{
			BudgetID:     budgetID,
			Date:         types.UnixTime{Time: entry.TransDate},
			AccountID:    id,
			PayeeID:      payeeID,
			TotalOutflow: entry.Outflow,
			TotalInflow:  entry.Inflow,
			Note:         entry.Note,
		})
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(strconv.Itoa(len(stmt.Entries))))
}

func resolveOrCreatePayeeID(ctx context.Context, queries *data.Queries, budgetID int64, payeeName string) (int64, error) {
	payee, err := queries.GetPayeeByName(ctx, data.GetPayeeByNameParams{
		BudgetID: budgetID,
		Name:     payeeName,
	})
	if err == nil {
		return payee.ID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	created, createErr := queries.CreatePayee(ctx, data.CreatePayeeParams{
		BudgetID: budgetID,
		Name:     payeeName,
	})
	if createErr != nil {
		return 0, createErr
	}

	return created.ID, nil
}
