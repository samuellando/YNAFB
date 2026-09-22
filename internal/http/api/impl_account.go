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
	"samuellando.com/YNAFB/internal/domain"
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
	accounts, err := s.accountService.List(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	// Send response
	resp := GetBudgetBudgetIdAccount200JSONResponse{}
	for _, account := range accounts {
		balance, err := account.Balance(ctx)
		if err != nil {
			return nil, err
		}
		reconciledBalance, err := account.ReconciledBalance(ctx)
		if err != nil {
			return nil, err
		}
		resp = append(resp, AccountSummaryDetail{
			Id:                account.ID(),
			Name:              account.Name(),
			Balance:           balance,
			ReconciledBalance: reconciledBalance,
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
	account, err := s.accountService.Create(ctx, int(loginID), budgetID, name)
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdAccount200JSONResponse{
		Id:   account.ID(),
		Name: account.Name(),
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
	account, err := s.accountService.Get(ctx, int(loginID), budgetID, id) 
	if err != nil {
		return nil, err
	}
	transactions, err := account.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	// Send response
	balance, err := account.Balance(ctx)
	if err != nil {
		return nil, err
	}
	reconciledBalance, err := account.ReconciledBalance(ctx)
	if err != nil {
		return nil, err
	}
	return GetBudgetBudgetIdAccountId200JSONResponse{
		Summary: AccountSummaryDetail{
			Id:                account.ID(),
			Name:              account.Name(),
			Balance:           balance,
			ReconciledBalance: reconciledBalance,
		},
		Transactions: marshalTrxs(transactions),
	}, nil
}

func marshalTrxs(trxs []*domain.Trx) []AccountTransaction {
	res := make([]AccountTransaction, 0)
	for _, trx := range trxs {
		accountTransaction := AccountTransaction{
			Id: trx.ID(),
			Date: trx.Date().Format(time.RFC3339),
			PayeeId: trx.Payee().ID(),
			PayeeName: trx.Payee().Name(),

			Outflow: trx.TotalOutflow(),
			Inflow: trx.TotalInflow(),

			TransactionLines: marshalTrxLines(trx.Lines()),

			Note: trx.Note(),
			Reconciled: trx.Reconciled(),
		}
		res = append(res, accountTransaction)
	}
	return res
}

func marshalTrxLines(lines []*domain.TrxLine) []AccountTransactionLine {
	res := make([]AccountTransactionLine, 0)
	for _, line := range lines {
		accountTransactionLine := AccountTransactionLine{
			LineId: line.ID(),
			Inflow: line.Inflow(),
			Outflow: line.Outflow(),
			Income: line.IsIncome(),
		}
		if category, err := line.Category(); err == nil {
			id := category.ID()
			name := category.Name()
			accountTransactionLine.CategoryId = &id
			accountTransactionLine.CategoryName = &name
		}
		if destAccount, err := line.DestinationAccount(); err == nil {
			id := destAccount.ID()
			name := destAccount.Name()
			accountTransactionLine.DestAccountId = &id
			accountTransactionLine.DestAccountName = &name
		}
		res = append(res, accountTransactionLine)
	}
	return res
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
	account, err := s.accountService.Get(ctx, int(loginID), budgetID, id) 
	if err != nil {
		return nil, err
	}
	err = account.Update(ctx, name)
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdAccountId200JSONResponse{
		Id:   account.ID(),
		Name: account.Name(),
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
	account, err := s.accountService.Get(ctx, int(loginID), budgetID, id) 
	if err != nil {
		return nil, err
	}
	err = account.Delete(ctx)
	if err != nil {
		return nil, err
	}
	// Send the response
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
