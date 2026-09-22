package api

import (
	"context"
	"fmt"
	"strconv"
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
	lines, err := s.payeeService.ListDefaultLinesByPayee(ctx, int(loginID), budgetID, payeeID)
	if err != nil {
		return nil, err
	}
	// Send response
	resp := GetBudgetBudgetIdPayeePayeeIdDefaultLine200JSONResponse{}
	for _, line := range lines {
		resp = append(resp, PayeeDefaultLine{
			Id:              int(line.ID),
			PayeeId:         int(line.PayeeID),
			DestAccountId:   nullInt64ToInt(line.DestAccountID),
			DestAccountName: nullStringToPtr(line.DestAccountName),
			CategoryId:      nullInt64ToInt(line.CategoryID),
			CategoryName:    nullStringToPtr(line.CategoryName),
			Income:          line.Income,
			Percent:         int(line.Percent),
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
	line, err := s.payeeService.CreateDefaultLine(ctx, int(loginID), budgetID, payeeID, request.Body.DestAccountId, request.Body.CategoryId, BoolPtrToBool(request.Body.Income), request.Body.Percent)
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdPayeePayeeIdDefaultLine200JSONResponse{
		Id:            line.ID(),
		PayeeId:       line.PayeeID(),
		DestAccountId: line.DestAccountID(),
		CategoryId:    line.CategoryID(),
		Income:        line.Income(),
		Percent:       line.Percent(),
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
	line, err := s.payeeService.GetDefaultLine(ctx, int(loginID), budgetID, id)
	if err != nil {
		return nil, err
	}
	err = line.Update(ctx, int(loginID), payeeID, request.Body.DestAccountId, request.Body.CategoryId, BoolPtrToBool(request.Body.Income), request.Body.Percent)
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdPayeePayeeIdDefaultLineId200JSONResponse{
		Id:            line.ID(),
		PayeeId:       line.PayeeID(),
		DestAccountId: line.DestAccountID(),
		CategoryId:    line.CategoryID(),
		Income:        line.Income(),
		Percent:       line.Percent(),
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
	line, err := s.payeeService.GetDefaultLine(ctx, int(loginID), budgetID, id)
	if err != nil {
		return nil, err
	}
	err = line.Delete(ctx, int(loginID))
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdPayeePayeeIdDefaultLineId200Response{}, nil
}
