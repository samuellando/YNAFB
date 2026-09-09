package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db/types"
	"samuellando.com/YNAFB/internal/importer"
)

func (s ApiServer) GetBudgetBudgetIdAccount(ctx context.Context, request GetBudgetBudgetIdAccountRequestObject) (GetBudgetBudgetIdAccountResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetString := request.BudgetId
	budgetID, err := strconv.Atoi(budgetString)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	// Query the database
	accounts, err := s.queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{
		LoginID:  loginID,
		BudgetID: int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	// Send response
	resp := GetBudgetBudgetIdAccount200JSONResponse{}
	for _, account := range accounts {
		resp = append(resp, AccountSummaryDetail{
			Id:                int(account.ID),
			Name:              account.Name,
			Balance:           int(account.Balance),
			ReconciledBalance: int(account.ReconciledBalance),
		})
	}
	return resp, nil
}

func (s ApiServer) PostBudgetBudgetIdAccount(ctx context.Context, request PostBudgetBudgetIdAccountRequestObject) (PostBudgetBudgetIdAccountResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetString := request.BudgetId
	budgetID, err := strconv.Atoi(budgetString)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`name` is required in request body")
	}
	name := request.Body.Name
	// Query the database
	account, err := s.queries.CreateAccount(ctx, data.CreateAccountParams{
		Name:     name,
		LoginID:  loginID,
		BudgetID: int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdAccount200JSONResponse{
		Id:   int(account.ID),
		Name: account.Name,
	}, nil
}

func (s ApiServer) GetBudgetBudgetIdAccountId(ctx context.Context, request GetBudgetBudgetIdAccountIdRequestObject) (GetBudgetBudgetIdAccountIdResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetString := request.BudgetId
	budgetID, err := strconv.Atoi(budgetString)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid account id: %w", err)
	}
	// Query the database
	summary, err := s.queries.GetAccountBalances(ctx, data.GetAccountBalancesParams{
		LoginID:  loginID,
		BudgetID: int64(budgetID),
		ID:       int64(id),
	})
	if err != nil {
		return nil, err
	}
	transactions, err := s.queries.ListAccountTransactions(ctx, data.ListAccountTransactionsParams{
		LoginID:  loginID,
		BudgetID: int64(budgetID),
		ID:       int64(id),
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return GetBudgetBudgetIdAccountId200JSONResponse{
		Summary: AccountSummaryDetail{
			Id:                int(summary.ID),
			Name:              summary.Name,
			Balance:           int(summary.Balance),
			ReconciledBalance: int(summary.ReconciledBalance),
		},
		Transactions: groupAccountTransactions(transactions),
	}, nil
}

func groupAccountTransactions(rows []data.ListAccountTransactionsRow) []AccountTransaction {
	transactions := make([]AccountTransaction, 0)
	for _, row := range rows {
		if len(transactions) == 0 || transactions[len(transactions)-1].Id != int(row.TrxID) {
			var sourceAccountID *int
			if row.SourceAccountID.Valid {
				sourceID := int(row.SourceAccountID.Int64)
				sourceAccountID = &sourceID
			}
			var sourceAccountName *string
			if row.SourceAccountName.Valid {
				name := row.SourceAccountName.String
				sourceAccountName = &name
			}
			var payeeName string
			if row.PayeeName.Valid {
				payeeName = row.PayeeName.String
			}
			transactions = append(transactions, AccountTransaction{
				Id:                int(row.TrxID),
				Date:              row.Date.Format(time.RFC3339),
				PayeeId:           int(row.PayeeID),
				PayeeName:         payeeName,
				Outflow:           int(row.Outflow),
				Inflow:            int(row.Inflow),
				Note:              row.Note,
				Reconciled:        row.Reconciled,
				SourceAccountId:   sourceAccountID,
				SourceAccountName: sourceAccountName,
				TransactionLines:  make([]AccountTransactionLine, 0),
			})
		}
		tx := &transactions[len(transactions)-1]
		if !row.TrxLineID.Valid {
			continue
		}
		var destAccountID *int
		if row.DestAccountID.Valid {
			destID := int(row.DestAccountID.Int64)
			destAccountID = &destID
		}
		var destAccountName *string
		if row.DestAccountName.Valid {
			name := row.DestAccountName.String
			destAccountName = &name
		}
		var categoryID *int
		if row.CategoryID.Valid {
			catID := int(row.CategoryID.Int64)
			categoryID = &catID
		}
		var categoryName *string
		if row.CategoryName.Valid {
			name := row.CategoryName.String
			categoryName = &name
		}
		lineOutflow := 0
		if row.LineOutflow.Valid {
			lineOutflow = int(row.LineOutflow.Int64)
		}
		lineInflow := 0
		if row.LineInflow.Valid {
			lineInflow = int(row.LineInflow.Int64)
		}
		tx.TransactionLines = append(tx.TransactionLines, AccountTransactionLine{
			LineId:          int(row.TrxLineID.Int64),
			DestAccountId:   destAccountID,
			DestAccountName: destAccountName,
			CategoryId:      categoryID,
			CategoryName:    categoryName,
			Income:          row.Income,
			Outflow:         lineOutflow,
			Inflow:          lineInflow,
		})
	}
	return transactions
}

func (s ApiServer) PutBudgetBudgetIdAccountId(ctx context.Context, request PutBudgetBudgetIdAccountIdRequestObject) (PutBudgetBudgetIdAccountIdResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetString := request.BudgetId
	budgetID, err := strconv.Atoi(budgetString)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid account id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`name` is required in request body")
	}
	name := request.Body.Name
	// Query the database
	account, err := s.queries.UpdateAccount(ctx, data.UpdateAccountParams{
		Name:     name,
		ID:       int64(id),
		LoginID:  loginID,
		BudgetID: int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdAccountId200JSONResponse{
		Id:   int(account.ID),
		Name: account.Name,
	}, nil
}

func (s ApiServer) DeleteBudgetBudgetIdAccountId(ctx context.Context, request DeleteBudgetBudgetIdAccountIdRequestObject) (DeleteBudgetBudgetIdAccountIdResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetString := request.BudgetId
	budgetID, err := strconv.Atoi(budgetString)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid account id: %w", err)
	}
	// Query the database
	err = s.queries.DeleteAccount(ctx, data.DeleteAccountParams{
		ID:       int64(id),
		LoginID:  loginID,
		BudgetID: int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdAccountId200Response{}, nil
}

func (s ApiServer) PostBudgetBudgetIdAccountIdReconcile(ctx context.Context, request PostBudgetBudgetIdAccountIdReconcileRequestObject) (PostBudgetBudgetIdAccountIdReconcileResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetString := request.BudgetId
	budgetID, err := strconv.Atoi(budgetString)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid account id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`date` and `balance` are required in request body")
	}
	date, err := time.Parse(time.RFC3339, request.Body.Date)
	if err != nil {
		return nil, fmt.Errorf("Invalid date: %w", err)
	}
	// Query the database
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	txQueries := s.queries.WithTx(tx)

	balance, err := txQueries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		LoginID:  loginID,
		BudgetID: int64(budgetID),
		ID:       int64(id),
		Date:     types.UnixTime{Time: date},
	})
	if err != nil {
		return nil, err
	}
	if int(balance) != request.Body.Balance {
		return nil, fmt.Errorf("balance mismatch: statement %d != calculated %d", request.Body.Balance, balance)
	}

	_, err = txQueries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID: int64(budgetID),
		ID:       int64(id),
		LoginID:  loginID,
		Date:     types.UnixTime{Time: date},
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return PostBudgetBudgetIdAccountIdReconcile200Response{}, nil
}

func (s ApiServer) PostBudgetBudgetIdAccountIdImport(ctx context.Context, request PostBudgetBudgetIdAccountIdImportRequestObject) (PostBudgetBudgetIdAccountIdImportResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetString := request.BudgetId
	budgetID, err := strconv.Atoi(budgetString)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid account id: %w", err)
	}
	// Read the statement file from the multipart body
	var content []byte
	for {
		part, err := request.Body.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if part.FormName() == "statement" {
			content, err = io.ReadAll(part)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	if content == nil {
		return nil, fmt.Errorf("`statement` is required in request body")
	}
	stmt, err := importer.Parse(content)
	if err != nil {
		return nil, err
	}
	// Query the database
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	txQueries := s.queries.WithTx(tx)
	for _, entry := range stmt.Entries {
		payeeName := strings.TrimSpace(entry.Payee)
		if payeeName == "" {
			payeeName = "unknown"
		}

		payeeID, err := resolveOrCreatePayeeID(ctx, txQueries, loginID, int64(budgetID), payeeName)
		if err != nil {
			return nil, err
		}

		_, err = txQueries.CreateTrx(ctx, data.CreateTrxParams{
			LoginID:      loginID,
			BudgetID:     int64(budgetID),
			Date:         types.UnixTime{Time: entry.TransDate},
			AccountID:    int64(id),
			PayeeID:      payeeID,
			TotalOutflow: entry.Outflow,
			TotalInflow:  entry.Inflow,
			Note:         entry.Note,
		})
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdAccountIdImport200TextResponse(len(stmt.Entries)), nil
}

func resolveOrCreatePayeeID(ctx context.Context, queries *data.Queries, loginID, budgetID int64, payeeName string) (int64, error) {
	payee, err := queries.GetPayeeByName(ctx, data.GetPayeeByNameParams{
		BudgetID: budgetID,
		LoginID:  loginID,
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
		LoginID:  loginID,
		Name:     payeeName,
	})
	if createErr != nil {
		return 0, createErr
	}

	return created.ID, nil
}
