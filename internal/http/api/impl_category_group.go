package api

import (
	"context"
	"fmt"
	"strconv"
)

func (s ApiServer) GetBudgetBudgetIdCategoryGroup(ctx context.Context, request GetBudgetBudgetIdCategoryGroupRequestObject) (GetBudgetBudgetIdCategoryGroupResponseObject, error) {
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
	groups, err := s.categoryService.ListGroups(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	// Send response
	resp := GetBudgetBudgetIdCategoryGroup200JSONResponse{}
	for _, group := range groups {
		resp = append(resp, CategoryGroup{
			Id:   group.ID(),
			Name: group.Name(),
		})
	}
	return resp, nil
}

func (s ApiServer) PostBudgetBudgetIdCategoryGroup(ctx context.Context, request PostBudgetBudgetIdCategoryGroupRequestObject) (PostBudgetBudgetIdCategoryGroupResponseObject, error) {
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
	group, err := s.categoryService.CreateGroup(ctx, int(loginID), budgetID, request.Body.Name)
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdCategoryGroup200JSONResponse{
		Id:   group.ID(),
		Name: group.Name(),
	}, nil
}

func (s ApiServer) PutBudgetBudgetIdCategoryGroupId(ctx context.Context, request PutBudgetBudgetIdCategoryGroupIdRequestObject) (PutBudgetBudgetIdCategoryGroupIdResponseObject, error) {
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
		return nil, fmt.Errorf("Invalid category group id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`name` is required in request body")
	}
	// Query the database
	group, err := s.categoryService.GetGroup(ctx, int(loginID), budgetID, id)
	if err != nil {
		return nil, err
	}
	err = group.Update(ctx, int(loginID), request.Body.Name)
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdCategoryGroupId200JSONResponse{
		Id:   group.ID(),
		Name: group.Name(),
	}, nil
}

func (s ApiServer) DeleteBudgetBudgetIdCategoryGroupId(ctx context.Context, request DeleteBudgetBudgetIdCategoryGroupIdRequestObject) (DeleteBudgetBudgetIdCategoryGroupIdResponseObject, error) {
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
		return nil, fmt.Errorf("Invalid category group id: %w", err)
	}
	// Query the database
	group, err := s.categoryService.GetGroup(ctx, int(loginID), budgetID, id)
	if err != nil {
		return nil, err
	}
	err = group.Delete(ctx, int(loginID))
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdCategoryGroupId200Response{}, nil
}