package api

import (
	"context"
	"fmt"
	"strconv"

	"samuellando.com/YNAFB/data"
)

func (s ApiServer) PostBudgetBudgetIdAccountAccountIdTransactionTrxIdLine(ctx context.Context, request PostBudgetBudgetIdAccountAccountIdTransactionTrxIdLineRequestObject) (PostBudgetBudgetIdAccountAccountIdTransactionTrxIdLineResponseObject, error) {
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
	trxString := request.TrxId
	trxID, err := strconv.Atoi(trxString)
	if err != nil {
		return nil, fmt.Errorf("Invalid transaction id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`outflow` and `inflow` are required in request body")
	}
	// Query the database
	line, err := s.queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
		TrxID:         int64(trxID),
		DestAccountID: intToNullInt64(request.Body.DestAccountId),
		CategoryID:    intToNullInt64(request.Body.CategoryId),
		Income:        BoolPtrToBool(request.Body.Income),
		Outflow:       int64(request.Body.Outflow),
		Inflow:        int64(request.Body.Inflow),
		BudgetID:      int64(budgetID),
		LoginID:       loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdAccountAccountIdTransactionTrxIdLine200JSONResponse{
		Id:            int(line.ID),
		TrxId:         int(line.TrxID),
		DestAccountId: nullInt64ToInt(line.DestAccountID),
		CategoryId:    nullInt64ToInt(line.CategoryID),
		Income:        line.Income,
		Outflow:       int(line.Outflow),
		Inflow:        int(line.Inflow),
	}, nil
}

func (s ApiServer) PutBudgetBudgetIdAccountAccountIdTransactionTrxIdLineId(ctx context.Context, request PutBudgetBudgetIdAccountAccountIdTransactionTrxIdLineIdRequestObject) (PutBudgetBudgetIdAccountAccountIdTransactionTrxIdLineIdResponseObject, error) {
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
	trxString := request.TrxId
	trxID, err := strconv.Atoi(trxString)
	if err != nil {
		return nil, fmt.Errorf("Invalid transaction id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid line id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`outflow` and `inflow` are required in request body")
	}
	// Query the database
	line, err := s.queries.UpdateTrxLine(ctx, data.UpdateTrxLineParams{
		TrxID:         int64(trxID),
		DestAccountID: intToNullInt64(request.Body.DestAccountId),
		CategoryID:    intToNullInt64(request.Body.CategoryId),
		Income:        BoolPtrToBool(request.Body.Income),
		Outflow:       int64(request.Body.Outflow),
		Inflow:        int64(request.Body.Inflow),
		ID:            int64(id),
		BudgetID:      int64(budgetID),
		LoginID:       loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdAccountAccountIdTransactionTrxIdLineId200JSONResponse{
		Id:            int(line.ID),
		TrxId:         int(line.TrxID),
		DestAccountId: nullInt64ToInt(line.DestAccountID),
		CategoryId:    nullInt64ToInt(line.CategoryID),
		Income:        line.Income,
		Outflow:       int(line.Outflow),
		Inflow:        int(line.Inflow),
	}, nil
}

func (s ApiServer) DeleteBudgetBudgetIdAccountAccountIdTransactionTrxIdLineId(ctx context.Context, request DeleteBudgetBudgetIdAccountAccountIdTransactionTrxIdLineIdRequestObject) (DeleteBudgetBudgetIdAccountAccountIdTransactionTrxIdLineIdResponseObject, error) {
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
	trxString := request.TrxId
	_, err = strconv.Atoi(trxString)
	if err != nil {
		return nil, fmt.Errorf("Invalid transaction id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid line id: %w", err)
	}
	// Query the database
	err = s.queries.DeleteTrxLine(ctx, data.DeleteTrxLineParams{
		ID:       int64(id),
		BudgetID: int64(budgetID),
		LoginID:  loginID,
	})
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdAccountAccountIdTransactionTrxIdLineId200Response{}, nil
}