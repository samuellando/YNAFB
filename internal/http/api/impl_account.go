package api

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"samuellando.com/YNAFB/internal/domain"
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
	account, err := s.accountService.Get(ctx, int(loginID), budgetID, id)
	if err != nil {
		return nil, err
	}
	if err := account.Reconcile(ctx, date, request.Body.Balance); err != nil {
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
	// Query the database inside a transaction so the import is atomic.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	txAccountService := s.accountService.WithRepo(s.queries.WithTx(tx))
	account, err := txAccountService.Get(ctx, int(loginID), budgetID, id)
	if err != nil {
		return nil, err
	}
	n, err := account.ImportStatement(ctx, stmt)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdAccountIdImport200TextResponse(n), nil
}
