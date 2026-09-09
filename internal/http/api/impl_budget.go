package api

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db/types"
)

type ApiServer struct {
	queries *data.Queries
	db      *sql.DB
}

var _ StrictServerInterface = (*ApiServer)(nil)

func NewServer(db *sql.DB) ApiServer {
	return ApiServer{
		queries: data.New(db),
		db:      db,
	}
}

func (s ApiServer) GetBudget(ctx context.Context, request GetBudgetRequestObject) (GetBudgetResponseObject, error) {
	// Collect params
	id, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	// Query the DB
	params := data.ListBudgetsParams{
		LoginID: id,
	}
	budgets, err := s.queries.ListBudgets(ctx, params)
	if err != nil {
		return nil, err
	}
	// Send response
	resp := GetBudget200JSONResponse{}
	for _, budget := range budgets {
		resp = append(resp, struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		}{
			Id:   int(budget.ID),
			Name: budget.Name,
		})
	}
	return resp, nil
}

func (s ApiServer) PostBudget(ctx context.Context, request PostBudgetRequestObject) (PostBudgetResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	nameAny, ok := (*request.Body)["name"]
	if !ok {
		return nil, fmt.Errorf("`name` is required in request body")
	}
	name, ok := nameAny.(string)
	if !ok {
		return nil, fmt.Errorf("`name` must be of type string")
	}
	// Query the DB
	budget, err := s.queries.CreateBudget(ctx, data.CreateBudgetParams{
		LoginID: loginID,
		Name:    name,
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudget201JSONResponse{
		Id:   int(budget.ID),
		Name: budget.Name,
	}, nil
}

func (s ApiServer) GetBudgetBudgetId(ctx context.Context, request GetBudgetBudgetIdRequestObject) (GetBudgetBudgetIdResponseObject, error) {
	now := time.Now().Format("2006-01")
	resp, err := s.getBudgetMonth(ctx, GetBudgetBudgetIdMonthRequestObject{
		BudgetId: request.BudgetId,
		Month:    now,
	})
	if err != nil {
		return nil, err
	}
	return GetBudgetBudgetId200JSONResponse(*resp), nil
}

func (s ApiServer) GetBudgetBudgetIdMonth(ctx context.Context, request GetBudgetBudgetIdMonthRequestObject) (GetBudgetBudgetIdMonthResponseObject, error) {
	resp, err := s.getBudgetMonth(ctx, request)
	if err != nil {
		return nil, err
	}
	return GetBudgetBudgetIdMonth200JSONResponse(*resp), nil
}

func (s ApiServer) getBudgetMonth(ctx context.Context, request GetBudgetBudgetIdMonthRequestObject) (*BudgetMonth, error) {
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
	monthString := request.Month
	month, err := time.Parse("2006-01", monthString)
	if err != nil {
		return nil, fmt.Errorf("Invalid month: %w", err)
	}
	// Query the database
	summary, err := s.queries.GetBudgetMonthSummary(ctx, data.GetBudgetMonthSummaryParams{
		LoginID: loginID,
		Month:   types.UnixTime{Time: month},
		ID:      int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	categories, err := s.queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
		LoginID: loginID,
		Month:   types.UnixTime{Time: month},
		ID:      int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	goals, err := s.queries.ListGoalsValues(ctx, data.ListGoalsValuesParams{
		LoginID: loginID,
		Month:   types.UnixTime{Time: month},
		ID:      int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	// Send response
	goalsMap := make(map[int64]data.ListGoalsValuesRow)
	for _, goal := range goals {
		goalsMap[goal.CategoryID] = goal
	}
	respCategories := make([]BudgetMonthCategory, 0)
	for _, category := range categories {
		categoryGroupName := &category.CategoryGroupName.String
		if !category.CategoryGroupName.Valid {
			categoryGroupName = nil
		}
		var goal *BudgetMonthGoal = nil
		if goalRow, ok := goalsMap[category.ID]; ok {
			var endMonth *string = nil
			if goalRow.EndDate.Valid {
				endMonthString := goalRow.EndDate.Time.Format(time.RFC3339)
				endMonth = &endMonthString
			}
			goal = &BudgetMonthGoal{
				Allocated:      int(goalRow.Allocated),
				Amount:         int(goalRow.Amount),
				AmountForMonth: int(goalRow.AmountForMonth),
				Gap:            int(goalRow.Gap),
				StartMonth:     goalRow.StartDate.Format(time.RFC3339),
				EndMonth:       endMonth,
			}
		}
		respCategories = append(respCategories, BudgetMonthCategory{
			CategoryId:        int(category.ID),
			CategoryName:      category.CategoryName,
			CategoryGroupName: categoryGroupName,
			Allocated:         int(category.Allocated),
			Available:         int(category.Available),
			Spent:             int(category.Spent),
			Goal:              goal,
		})
	}
	return &BudgetMonth{
		Month: month.Format(time.RFC3339),
		Summary: BudgetMonthSummary{
			Allocated:     int(summary.Allocated),
			Available:     int(summary.Available),
			Income:        int(summary.Income),
			ReadyToAssign: int(summary.ReadyToAssign),
			Spent:         int(summary.Spent),
			Uncategorized: int(summary.Uncategorized),
		},
		Categories: respCategories,
	}, nil
}

func (s ApiServer) PutBudgetBudgetId(ctx context.Context, request PutBudgetBudgetIdRequestObject) (PutBudgetBudgetIdResponseObject, error) {
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
	nameAny, ok := (*request.Body)["name"]
	if !ok {
		return nil, fmt.Errorf("`name` is required in request body")
	}
	name, ok := nameAny.(string)
	if !ok {
		return nil, fmt.Errorf("`name` must be of type string")
	}
	// Query the DB
	budget, err := s.queries.UpdateBudget(ctx, data.UpdateBudgetParams{
		LoginID: loginID,
		ID:      int64(budgetID),
		Name:    name,
	})
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetId200JSONResponse{
		Id:   int(budget.ID),
		Name: budget.Name,
	}, nil
}

func (s ApiServer) DeleteBudgetBudgetId(ctx context.Context, request DeleteBudgetBudgetIdRequestObject) (DeleteBudgetBudgetIdResponseObject, error) {
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
	// Query the DB
	err = s.queries.DeleteBudget(ctx, data.DeleteBudgetParams{
		LoginID: loginID,
		ID:      int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}
