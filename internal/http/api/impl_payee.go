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
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	payees, err := budget.ListPayees(ctx)
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
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	payee, err := budget.CreatePayee(ctx, request.Body.Name)
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
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	payee, err := budget.GetPayee(ctx, id)
	if err != nil {
		return nil, err
	}
	err = payee.Update(ctx, request.Body.Name)
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
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	payee, err := budget.GetPayee(ctx, id)
	if err != nil {
		return nil, err
	}
	err = payee.Delete(ctx)
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdPayeeId200Response{}, nil
}