package api

import (
	"context"
	"fmt"
	"strconv"

	"samuellando.com/YNAFB/data"
)

func (s ApiServer) GetBudgetBudgetIdPayeePayeeIdDefaultLine(ctx context.Context, request GetBudgetBudgetIdPayeePayeeIdDefaultLineRequestObject) (GetBudgetBudgetIdPayeePayeeIdDefaultLineResponseObject, error) {
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
	payeeString := request.PayeeId
	payeeID, err := strconv.Atoi(payeeString)
	if err != nil {
		return nil, fmt.Errorf("Invalid payee id: %w", err)
	}
	// Query the database
	lines, err := s.queries.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{
		LoginID:  loginID,
		PayeeID:  int64(payeeID),
		BudgetID: int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	// Send response
	resp := GetBudgetBudgetIdPayeePayeeIdDefaultLine200JSONResponse{}
	for _, line := range lines {
		resp = append(resp, PayeeDefaultLine{
			Id:               int(line.ID),
			PayeeId:          int(line.PayeeID),
			DestAccountId:    nullInt64ToInt(line.DestAccountID),
			DestAccountName:  nullStringToPtr(line.DestAccountName),
			CategoryId:       nullInt64ToInt(line.CategoryID),
			CategoryName:     nullStringToPtr(line.CategoryName),
			ExpenseShareId:   nullInt64ToInt(line.ExpenseShareID),
			ExpenseShareName: nullStringToPtr(line.ExpenseShareName),
			Income:           line.Income,
			Percent:          int(line.Percent),
		})
	}
	return resp, nil
}

func (s ApiServer) PostBudgetBudgetIdPayeePayeeIdDefaultLine(ctx context.Context, request PostBudgetBudgetIdPayeePayeeIdDefaultLineRequestObject) (PostBudgetBudgetIdPayeePayeeIdDefaultLineResponseObject, error) {
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
	payeeString := request.PayeeId
	payeeID, err := strconv.Atoi(payeeString)
	if err != nil {
		return nil, fmt.Errorf("Invalid payee id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`percent` is required in request body")
	}
	// Query the database
	line, err := s.queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		PayeeID:       int64(payeeID),
		DestAccountID: intToNullInt64(request.Body.DestAccountId),
		CategoryID:    intToNullInt64(request.Body.CategoryId),
		ExpenseShareID:    intToNullInt64(request.Body.ExpenseShareId),
		Income:        BoolPtrToBool(request.Body.Income),
		Percent:       int64(request.Body.Percent),
		BudgetID:      int64(budgetID),
		LoginID:       loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdPayeePayeeIdDefaultLine200JSONResponse{
		Id:            int(line.ID),
		PayeeId:       int(line.PayeeID),
		DestAccountId: nullInt64ToInt(line.DestAccountID),
		CategoryId:    nullInt64ToInt(line.CategoryID),
		ExpenseShareId:    nullInt64ToInt(line.ExpenseShareID),
		Income:        line.Income,
		Percent:       int(line.Percent),
	}, nil
}

func (s ApiServer) PutBudgetBudgetIdPayeePayeeIdDefaultLineId(ctx context.Context, request PutBudgetBudgetIdPayeePayeeIdDefaultLineIdRequestObject) (PutBudgetBudgetIdPayeePayeeIdDefaultLineIdResponseObject, error) {
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
	payeeString := request.PayeeId
	payeeID, err := strconv.Atoi(payeeString)
	if err != nil {
		return nil, fmt.Errorf("Invalid payee id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid default line id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`percent` is required in request body")
	}
	// Query the database
	line, err := s.queries.UpdatePayeeDefaultLine(ctx, data.UpdatePayeeDefaultLineParams{
		PayeeID:       int64(payeeID),
		DestAccountID: intToNullInt64(request.Body.DestAccountId),
		CategoryID:    intToNullInt64(request.Body.CategoryId),
		ExpenseShareID:    intToNullInt64(request.Body.ExpenseShareId),
		Income:        BoolPtrToBool(request.Body.Income),
		Percent:       int64(request.Body.Percent),
		ID:            int64(id),
		BudgetID:      int64(budgetID),
		LoginID:       loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdPayeePayeeIdDefaultLineId200JSONResponse{
		Id:            int(line.ID),
		PayeeId:       int(line.PayeeID),
		DestAccountId: nullInt64ToInt(line.DestAccountID),
		CategoryId:    nullInt64ToInt(line.CategoryID),
		ExpenseShareId:    nullInt64ToInt(line.ExpenseShareID),
		Income:        line.Income,
		Percent:       int(line.Percent),
	}, nil
}

func (s ApiServer) DeleteBudgetBudgetIdPayeePayeeIdDefaultLineId(ctx context.Context, request DeleteBudgetBudgetIdPayeePayeeIdDefaultLineIdRequestObject) (DeleteBudgetBudgetIdPayeePayeeIdDefaultLineIdResponseObject, error) {
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
	payeeString := request.PayeeId
	_, err = strconv.Atoi(payeeString)
	if err != nil {
		return nil, fmt.Errorf("Invalid payee id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid default line id: %w", err)
	}
	// Query the database
	err = s.queries.DeletePayeeDefaultLine(ctx, data.DeletePayeeDefaultLineParams{
		ID:       int64(id),
		BudgetID: int64(budgetID),
		LoginID:  loginID,
	})
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdPayeePayeeIdDefaultLineId200Response{}, nil
}
