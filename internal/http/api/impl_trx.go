package api

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

func (s ApiServer) PostBudgetBudgetIdAccountAccountIdTransaction(ctx context.Context, request PostBudgetBudgetIdAccountAccountIdTransactionRequestObject) (PostBudgetBudgetIdAccountAccountIdTransactionResponseObject, error) {
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
	accountString := request.AccountId
	accountID, err := strconv.Atoi(accountString)
	if err != nil {
		return nil, fmt.Errorf("Invalid account id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`payeeId`, `date`, `outflow`, and `inflow` are required in request body")
	}
	date, err := time.Parse(time.RFC3339, request.Body.Date)
	if err != nil {
		return nil, fmt.Errorf("Invalid date: %w", err)
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	account, err := budget.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	payee, err := budget.GetPayee(ctx, request.Body.PayeeId)
	if err != nil {
		return nil, err
	}
	trx, err := account.CreateTransaction(ctx, payee, date, request.Body.Outflow, request.Body.Inflow, StrPtrToStr(request.Body.Note))
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdAccountAccountIdTransaction200JSONResponse{
		Id:        trx.ID(),
		AccountId: trx.Account().ID(),
		PayeeId:   trx.Payee().ID(),
		Date:      trx.Date().Format(time.RFC3339),
		Outflow:   trx.TotalOutflow(),
		Inflow:    trx.TotalInflow(),
		Note:      trx.Note(),
	}, nil
}

func (s ApiServer) PutBudgetBudgetIdAccountAccountIdTransactionId(ctx context.Context, request PutBudgetBudgetIdAccountAccountIdTransactionIdRequestObject) (PutBudgetBudgetIdAccountAccountIdTransactionIdResponseObject, error) {
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
	accountString := request.AccountId
	accountID, err := strconv.Atoi(accountString)
	if err != nil {
		return nil, fmt.Errorf("Invalid account id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid transaction id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`payeeId`, `date`, `outflow`, and `inflow` are required in request body")
	}
	date, err := time.Parse(time.RFC3339, request.Body.Date)
	if err != nil {
		return nil, fmt.Errorf("Invalid date: %w", err)
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	account, err := budget.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	payee, err := budget.GetPayee(ctx, request.Body.PayeeId)
	if err != nil {
		return nil, err
	}
	trx, err := account.GetTransaction(ctx, id)
	if err != nil {
		return nil, err
	}
	err = trx.Update(ctx, payee, date, request.Body.Outflow, request.Body.Inflow, StrPtrToStr(request.Body.Note))
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdAccountAccountIdTransactionId200JSONResponse{
		Id:        trx.ID(),
		AccountId: trx.Account().ID(),
		PayeeId:   trx.Payee().ID(),
		Date:      trx.Date().Format(time.RFC3339),
		Outflow:   trx.TotalOutflow(),
		Inflow:    trx.TotalInflow(),
		Note:      trx.Note(),
	}, nil
}

func (s ApiServer) DeleteBudgetBudgetIdAccountAccountIdTransactionId(ctx context.Context, request DeleteBudgetBudgetIdAccountAccountIdTransactionIdRequestObject) (DeleteBudgetBudgetIdAccountAccountIdTransactionIdResponseObject, error) {
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
	accountString := request.AccountId
	accountID, err := strconv.Atoi(accountString)
	if err != nil {
		return nil, fmt.Errorf("Invalid account id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid transaction id: %w", err)
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	account, err := budget.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	trx, err := account.GetTransaction(ctx, id)
	if err != nil {
		return nil, err
	}
	err = trx.Delete(ctx)
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdAccountAccountIdTransactionId200Response{}, nil
}
