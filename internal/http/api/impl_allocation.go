package api

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

func (s ApiServer) PutBudgetBudgetIdCategoryIdAllocationMonth(ctx context.Context, request PutBudgetBudgetIdCategoryIdAllocationMonthRequestObject) (PutBudgetBudgetIdCategoryIdAllocationMonthResponseObject, error) {
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
	categoryID, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid category id: %w", err)
	}
	month, err := time.Parse("2006-01", request.Month)
	if err != nil {
		return nil, fmt.Errorf("Invalid month: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`amount` is required in request body")
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	category, err := budget.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	allocation, err := category.SetAllocation(ctx, month, request.Body.Amount)
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdCategoryIdAllocationMonth200JSONResponse{
		Id:         allocation.ID(),
		CategoryId: allocation.Category().ID(),
		Month:      allocation.Month().Format(time.RFC3339),
		Amount:     allocation.Amount(),
	}, nil
}