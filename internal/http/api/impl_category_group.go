package api

import (
	"context"
	"fmt"
	"strconv"

	"samuellando.com/YNAFB/data"
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
	groups, err := s.queries.ListCategoryGroups(ctx, data.ListCategoryGroupsParams{
		LoginID:  loginID,
		BudgetID: int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	// Send response
	resp := GetBudgetBudgetIdCategoryGroup200JSONResponse{}
	for _, group := range groups {
		resp = append(resp, CategoryGroup{
			Id:   int(group.ID),
			Name: group.Name,
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
	id, err := s.queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		Name:     request.Body.Name,
		BudgetID: int64(budgetID),
		LoginID:  loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdCategoryGroup200JSONResponse{
		Id:   int(id),
		Name: request.Body.Name,
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
	group, err := s.queries.UpdateCategoryGroup(ctx, data.UpdateCategoryGroupParams{
		Name:     request.Body.Name,
		ID:       int64(id),
		BudgetID: int64(budgetID),
		LoginID:  loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdCategoryGroupId200JSONResponse{
		Id:   int(group.ID),
		Name: group.Name,
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
	err = s.queries.DeleteCategoryGroup(ctx, data.DeleteCategoryGroupParams{
		ID:       int64(id),
		BudgetID: int64(budgetID),
		LoginID:  loginID,
	})
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdCategoryGroupId200Response{}, nil
}