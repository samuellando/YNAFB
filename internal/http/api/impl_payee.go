package api

import (
	"context"
	"fmt"
	"strconv"
)

func (s ApiServer) GetBudgetBudgetIdPayee(ctx context.Context, request GetBudgetBudgetIdPayeeRequestObject) (GetBudgetBudgetIdPayeeResponseObject, error) {
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
	payees, err := s.payeeService.List(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	// Send response
	resp := GetBudgetBudgetIdPayee200JSONResponse{}
	for _, payee := range payees {
		resp = append(resp, Payee{
			Id:   payee.ID(),
			Name: payee.Name(),
		})
	}
	return resp, nil
}

func (s ApiServer) PostBudgetBudgetIdPayee(ctx context.Context, request PostBudgetBudgetIdPayeeRequestObject) (PostBudgetBudgetIdPayeeResponseObject, error) {
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
	// Query the database
	payee, err := s.payeeService.Create(ctx, int(loginID), budgetID, request.Body.Name)
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdPayee200JSONResponse{
		Id:   payee.ID(),
		Name: payee.Name(),
	}, nil
}

func (s ApiServer) PutBudgetBudgetIdPayeeId(ctx context.Context, request PutBudgetBudgetIdPayeeIdRequestObject) (PutBudgetBudgetIdPayeeIdResponseObject, error) {
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
		return nil, fmt.Errorf("Invalid payee id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`name` is required in request body")
	}
	// Query the database
	payee, err := s.payeeService.Get(ctx, int(loginID), budgetID, id)
	if err != nil {
		return nil, err
	}
	err = payee.Update(ctx, int(loginID), request.Body.Name)
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdPayeeId200JSONResponse{
		Id:   payee.ID(),
		Name: payee.Name(),
	}, nil
}

func (s ApiServer) DeleteBudgetBudgetIdPayeeId(ctx context.Context, request DeleteBudgetBudgetIdPayeeIdRequestObject) (DeleteBudgetBudgetIdPayeeIdResponseObject, error) {
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
		return nil, fmt.Errorf("Invalid payee id: %w", err)
	}
	// Query the database
	payee, err := s.payeeService.Get(ctx, int(loginID), budgetID, id)
	if err != nil {
		return nil, err
	}
	err = payee.Delete(ctx, int(loginID))
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdPayeeId200Response{}, nil
}