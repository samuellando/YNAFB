package api

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"samuellando.com/YNAFB/data"
)

func categoryGroupIDToNull(id *int) sql.NullInt64 {
	if id == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*id), Valid: true}
}

func nullToCategoryGroupID(id sql.NullInt64) *int {
	if !id.Valid {
		return nil
	}
	categoryGroupID := int(id.Int64)
	return &categoryGroupID
}

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
	categories, err := s.queries.ListCategories(ctx, data.ListCategoriesParams{
		LoginID:  loginID,
		BudgetID: int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	// Send response
	resp := GetBudgetBudgetIdCategory200JSONResponse{}
	for _, category := range categories {
		resp = append(resp, Category{
			Id:   int(category.ID),
			Name: category.Name,
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
	category, err := s.queries.CreateCategory(ctx, data.CreateCategoryParams{
		Name:            request.Body.Name,
		CategoryGroupID: categoryGroupIDToNull(request.Body.CategoryGroupId),
		BudgetID:        int64(budgetID),
		LoginID:         loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdCategory200JSONResponse{
		Id:              int(category.ID),
		Name:            category.Name,
		CategoryGroupId: nullToCategoryGroupID(category.CategoryGroupID),
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
	category, err := s.queries.UpdateCategory(ctx, data.UpdateCategoryParams{
		Name:            request.Body.Name,
		CategoryGroupID: categoryGroupIDToNull(request.Body.CategoryGroupId),
		ID:              int64(id),
		BudgetID:        int64(budgetID),
		LoginID:         loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdCategoryId200JSONResponse{
		Id:              int(category.ID),
		Name:            category.Name,
		CategoryGroupId: nullToCategoryGroupID(category.CategoryGroupID),
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
	err = s.queries.DeleteCategory(ctx, data.DeleteCategoryParams{
		ID:       int64(id),
		BudgetID: int64(budgetID),
		LoginID:  loginID,
	})
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdCategoryId200Response{}, nil
}