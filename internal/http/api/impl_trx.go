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
	account, err := s.accountService.Get(ctx, int(loginID), budgetID, accountID)
	if err != nil {
		return nil, err
	}
	payee, err := s.payeeService.Get(ctx, int(loginID), budgetID, request.Body.PayeeId)
	if err != nil {
		return nil, err
	}
	trx, err := s.trxService.Create(ctx, account, payee, date, request.Body.Outflow, request.Body.Inflow, StrPtrToStr(request.Body.Note))
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
	trx, err := s.trxService.Get(ctx, int(loginID), budgetID, id)
	if err != nil {
		return nil, err
	}
	err = trx.Update(ctx, int(loginID), accountID, request.Body.PayeeId, date, request.Body.Outflow, request.Body.Inflow, StrPtrToStr(request.Body.Note))
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
	_, err = strconv.Atoi(accountString)
	if err != nil {
		return nil, fmt.Errorf("Invalid account id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid transaction id: %w", err)
	}
	// Query the database
	trx, err := s.trxService.Get(ctx, int(loginID), budgetID, id)
	if err != nil {
		return nil, err
	}
	err = trx.Delete(ctx, int(loginID))
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdAccountAccountIdTransactionId200Response{}, nil
}
