package api

import (
	"context"
	"fmt"
	"strconv"
)

func (s ApiServer) GetBudgetBudgetIdCategory(ctx context.Context, request GetBudgetBudgetIdCategoryRequestObject) (GetBudgetBudgetIdCategoryResponseObject, error) {
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
	categories, err := s.categoryService.List(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	// Send response
	resp := GetBudgetBudgetIdCategory200JSONResponse{}
	for _, category := range categories {
		resp = append(resp, Category{
			Id:   category.ID(),
			Name: category.Name(),
		})
	}
	return resp, nil
}

func (s ApiServer) PostBudgetBudgetIdCategory(ctx context.Context, request PostBudgetBudgetIdCategoryRequestObject) (PostBudgetBudgetIdCategoryResponseObject, error) {
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
	category, err := s.categoryService.Create(ctx, int(loginID), budgetID, request.Body.Name, request.Body.CategoryGroupId)
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdCategory200JSONResponse{
		Id:              category.ID(),
		Name:            category.Name(),
		CategoryGroupId: category.GroupID(),
	}, nil
}

func (s ApiServer) PutBudgetBudgetIdCategoryId(ctx context.Context, request PutBudgetBudgetIdCategoryIdRequestObject) (PutBudgetBudgetIdCategoryIdResponseObject, error) {
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
		return nil, fmt.Errorf("Invalid category id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`name` is required in request body")
	}
	// Query the database
	category, err := s.categoryService.Get(ctx, int(loginID), budgetID, id)
	if err != nil {
		return nil, err
	}
	err = category.Update(ctx, int(loginID), request.Body.Name, request.Body.CategoryGroupId)
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdCategoryId200JSONResponse{
		Id:              category.ID(),
		Name:            category.Name(),
		CategoryGroupId: category.GroupID(),
	}, nil
}

func (s ApiServer) DeleteBudgetBudgetIdCategoryId(ctx context.Context, request DeleteBudgetBudgetIdCategoryIdRequestObject) (DeleteBudgetBudgetIdCategoryIdResponseObject, error) {
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
		return nil, fmt.Errorf("Invalid category id: %w", err)
	}
	// Query the database
	category, err := s.categoryService.Get(ctx, int(loginID), budgetID, id)
	if err != nil {
		return nil, err
	}
	err = category.Delete(ctx, int(loginID))
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdCategoryId200Response{}, nil
}
