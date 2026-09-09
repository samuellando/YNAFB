package api

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db/types"
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
	trx, err := s.queries.CreateTrx(ctx, data.CreateTrxParams{
		AccountID:    int64(accountID),
		PayeeID:      int64(request.Body.PayeeId),
		Date:         types.UnixTime{Time: date},
		TotalOutflow: int64(request.Body.Outflow),
		TotalInflow:  int64(request.Body.Inflow),
		Note:         StrPtrToStr(request.Body.Note),
		BudgetID:     int64(budgetID),
		LoginID:      loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdAccountAccountIdTransaction200JSONResponse{
		Id:        int(trx.ID),
		AccountId: int(trx.AccountID),
		PayeeId:   int(trx.PayeeID),
		Date:      trx.Date.Format(time.RFC3339),
		Outflow:   int(trx.TotalOutflow),
		Inflow:    int(trx.TotalInflow),
		Note:      trx.Note,
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
	trx, err := s.queries.UpdateTrx(ctx, data.UpdateTrxParams{
		Date:         types.UnixTime{Time: date},
		AccountID:    int64(accountID),
		PayeeID:      int64(request.Body.PayeeId),
		TotalOutflow: int64(request.Body.Outflow),
		TotalInflow:  int64(request.Body.Inflow),
		Note:         StrPtrToStr(request.Body.Note),
		ID:           int64(id),
		BudgetID:     int64(budgetID),
		LoginID:      loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdAccountAccountIdTransactionId200JSONResponse{
		Id:        int(trx.ID),
		AccountId: int(trx.AccountID),
		PayeeId:   int(trx.PayeeID),
		Date:      trx.Date.Format(time.RFC3339),
		Outflow:   int(trx.TotalOutflow),
		Inflow:    int(trx.TotalInflow),
		Note:      trx.Note,
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
	err = s.queries.DeleteTrx(ctx, data.DeleteTrxParams{
		ID:       int64(id),
		BudgetID: int64(budgetID),
		LoginID:  loginID,
	})
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdAccountAccountIdTransactionId200Response{}, nil
}
