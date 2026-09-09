package api

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db/types"
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
	allocation, err := s.queries.SetAllocation(ctx, data.SetAllocationParams{
		CategoryID: int64(categoryID),
		Month:      types.UnixTime{Time: month},
		Amount:     int64(request.Body.Amount),
		BudgetID:   int64(budgetID),
		LoginID:    loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdCategoryIdAllocationMonth200JSONResponse{
		Id:         int(allocation.ID),
		CategoryId: int(allocation.CategoryID),
		Month:      allocation.Month.Format(time.RFC3339),
		Amount:     int(allocation.Amount),
	}, nil
}