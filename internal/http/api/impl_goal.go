package api

import (
	"context"
	"fmt"
	"strconv"
	"time"
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
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	category, err := budget.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	goal, err := category.GetGoal(ctx)
	if err != nil {
		return nil, err
	}
	// Send response
	var endMonth *string
	if end := goal.EndDate(); end != nil {
		s := end.Format(time.RFC3339)
		endMonth = &s
	}
	return GetBudgetBudgetIdCategoryCategoryIdGoal200JSONResponse{
		Id:         goal.ID(),
		CategoryId: goal.Category().ID(),
		Type:       GoalType(goal.Type()),
		StartMonth: goal.StartDate().Format(time.RFC3339),
		EndMonth:   endMonth,
		Amount:     goal.Amount(),
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
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	category, err := budget.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	var end *time.Time
	if endMonth.Valid {
		t := endMonth.Time
		end = &t
	}
	goal, err := category.Create(ctx, string(request.Body.Type), startMonth.Time, end, request.Body.Amount)
	if err != nil {
		return nil, err
	}
	// Send response
	var endMonthStr *string
	if end := goal.EndDate(); end != nil {
		s := end.Format(time.RFC3339)
		endMonthStr = &s
	}
	return PostBudgetBudgetIdCategoryCategoryIdGoal200JSONResponse{
		Id:         goal.ID(),
		CategoryId: goal.Category().ID(),
		Type:       GoalType(goal.Type()),
		StartMonth: goal.StartDate().Format(time.RFC3339),
		EndMonth:   endMonthStr,
		Amount:     goal.Amount(),
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
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	category, err := budget.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	goal, err := category.GetGoal(ctx)
	if err != nil {
		return nil, err
	}
	var end *time.Time
	if endMonth.Valid {
		t := endMonth.Time
		end = &t
	}
	err = goal.Update(ctx, string(request.Body.Type), startMonth.Time, end, request.Body.Amount)
	if err != nil {
		return nil, err
	}
	// Send response
	var endMonthStr *string
	if end := goal.EndDate(); end != nil {
		s := end.Format(time.RFC3339)
		endMonthStr = &s
	}
	return PutBudgetBudgetIdCategoryCategoryIdGoal200JSONResponse{
		Id:         goal.ID(),
		CategoryId: goal.Category().ID(),
		Type:       GoalType(goal.Type()),
		StartMonth: goal.StartDate().Format(time.RFC3339),
		EndMonth:   endMonthStr,
		Amount:     goal.Amount(),
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
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	category, err := budget.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	goal, err := category.GetGoal(ctx)
	if err != nil {
		return nil, err
	}
	err = goal.Delete(ctx)
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdCategoryCategoryIdGoal200Response{}, nil
}
