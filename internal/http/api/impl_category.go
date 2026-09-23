package api

import (
	"context"
	"fmt"
	"strconv"

	"samuellando.com/YNAFB/internal/domain"
)

// resolveGroup loads the category group for an optional wire ID.
func resolveGroup(ctx context.Context, budget *domain.Budget, groupID *int) (*domain.CategoryGroup, error) {
	if groupID == nil {
		return nil, nil
	}
	return budget.GetCategoryGroup(ctx, *groupID)
}

// categoryGroupID translates a Category's group object to the nullable wire ID.
// The entity exposes objects (not scalar FKs), so the conversion lives here
// in the API layer.
func categoryGroupID(category *domain.Category) *int {
	if group, err := category.Group(); err == nil {
		id := group.ID()
		return &id
	}
	return nil
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
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	categories, err := budget.ListCategories(ctx)
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
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	group, err := resolveGroup(ctx, budget, request.Body.CategoryGroupId)
	if err != nil {
		return nil, err
	}
	category, err := budget.CreateCategory(ctx, request.Body.Name, group)
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdCategory200JSONResponse{
		Id:              category.ID(),
		Name:            category.Name(),
		CategoryGroupId: categoryGroupID(category),
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
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	category, err := budget.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}
	group, err := resolveGroup(ctx, budget, request.Body.CategoryGroupId)
	if err != nil {
		return nil, err
	}
	err = category.Update(ctx, request.Body.Name, group)
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdCategoryId200JSONResponse{
		Id:              category.ID(),
		Name:            category.Name(),
		CategoryGroupId: categoryGroupID(category),
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
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	category, err := budget.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}
	err = category.Delete(ctx)
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdCategoryId200Response{}, nil
}
