package api

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"samuellando.com/YNAFB/data"
)

func (s ApiServer) GetBudgetBudgetIdCategoryCategoryIdGoal(ctx context.Context, request GetBudgetBudgetIdCategoryCategoryIdGoalRequestObject) (GetBudgetBudgetIdCategoryCategoryIdGoalResponseObject, error) {
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
	categoryString := request.CategoryId
	categoryID, err := strconv.Atoi(categoryString)
	if err != nil {
		return nil, fmt.Errorf("Invalid category id: %w", err)
	}
	// Query the database
	goal, err := s.queries.GetGoalByCategory(ctx, data.GetGoalByCategoryParams{
		LoginID:    loginID,
		BudgetID:   int64(budgetID),
		CategoryID: int64(categoryID),
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return GetBudgetBudgetIdCategoryCategoryIdGoal200JSONResponse{
		Id:         int(goal.ID),
		CategoryId: int(goal.CategoryID),
		Type:       GoalType(goal.Type),
		StartMonth: goal.StartDate.Format(time.RFC3339),
		EndMonth:   nullUnixTimeToMonthString(goal.EndDate),
		Amount:     int(goal.Amount),
	}, nil
}

func (s ApiServer) PostBudgetBudgetIdCategoryCategoryIdGoal(ctx context.Context, request PostBudgetBudgetIdCategoryCategoryIdGoalRequestObject) (PostBudgetBudgetIdCategoryCategoryIdGoalResponseObject, error) {
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
	categoryString := request.CategoryId
	categoryID, err := strconv.Atoi(categoryString)
	if err != nil {
		return nil, fmt.Errorf("Invalid category id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`type`, `startMonth`, and `amount` are required in request body")
	}
	startMonth, err := parseMonth(request.Body.StartMonth)
	if err != nil {
		return nil, fmt.Errorf("Invalid start month: %w", err)
	}
	endMonth, err := parseNullableMonth(request.Body.EndMonth)
	if err != nil {
		return nil, fmt.Errorf("Invalid end month: %w", err)
	}
	// Query the database
	goal, err := s.queries.CreateGoal(ctx, data.CreateGoalParams{
		Type:       string(request.Body.Type),
		StartDate:  startMonth,
		EndDate:    endMonth,
		CategoryID: int64(categoryID),
		Amount:     int64(request.Body.Amount),
		BudgetID:   int64(budgetID),
		LoginID:    loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdCategoryCategoryIdGoal200JSONResponse{
		Id:         int(goal.ID),
		CategoryId: int(goal.CategoryID),
		Type:       GoalType(goal.Type),
		StartMonth: goal.StartDate.Format(time.RFC3339),
		EndMonth:   nullUnixTimeToMonthString(goal.EndDate),
		Amount:     int(goal.Amount),
	}, nil
}

func (s ApiServer) PutBudgetBudgetIdCategoryCategoryIdGoal(ctx context.Context, request PutBudgetBudgetIdCategoryCategoryIdGoalRequestObject) (PutBudgetBudgetIdCategoryCategoryIdGoalResponseObject, error) {
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
	categoryString := request.CategoryId
	categoryID, err := strconv.Atoi(categoryString)
	if err != nil {
		return nil, fmt.Errorf("Invalid category id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`type`, `startMonth`, and `amount` are required in request body")
	}
	startMonth, err := parseMonth(request.Body.StartMonth)
	if err != nil {
		return nil, fmt.Errorf("Invalid start month: %w", err)
	}
	endMonth, err := parseNullableMonth(request.Body.EndMonth)
	if err != nil {
		return nil, fmt.Errorf("Invalid end month: %w", err)
	}
	// Query the database
	goal, err := s.queries.UpdateGoal(ctx, data.UpdateGoalParams{
		Type:       string(request.Body.Type),
		StartDate:  startMonth,
		EndDate:    endMonth,
		Amount:     int64(request.Body.Amount),
		BudgetID:   int64(budgetID),
		LoginID:    loginID,
		CategoryID: int64(categoryID),
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdCategoryCategoryIdGoal200JSONResponse{
		Id:         int(goal.ID),
		CategoryId: int(goal.CategoryID),
		Type:       GoalType(goal.Type),
		StartMonth: goal.StartDate.Format(time.RFC3339),
		EndMonth:   nullUnixTimeToMonthString(goal.EndDate),
		Amount:     int(goal.Amount),
	}, nil
}

func (s ApiServer) DeleteBudgetBudgetIdCategoryCategoryIdGoal(ctx context.Context, request DeleteBudgetBudgetIdCategoryCategoryIdGoalRequestObject) (DeleteBudgetBudgetIdCategoryCategoryIdGoalResponseObject, error) {
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
	categoryString := request.CategoryId
	categoryID, err := strconv.Atoi(categoryString)
	if err != nil {
		return nil, fmt.Errorf("Invalid category id: %w", err)
	}
	// Query the database
	err = s.queries.DeleteGoal(ctx, data.DeleteGoalParams{
		CategoryID: int64(categoryID),
		BudgetID:   int64(budgetID),
		LoginID:    loginID,
	})
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdCategoryCategoryIdGoal200Response{}, nil
}
