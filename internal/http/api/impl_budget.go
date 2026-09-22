package api

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/domain"
)

type ApiServer struct {
	budgetService *domain.BudgetService
	queries *data.Queries
	db      *sql.DB
}

var _ StrictServerInterface = (*ApiServer)(nil)

func NewServer(db *sql.DB) ApiServer {
	queries := data.New(db)
	allocationService := domain.NewAllocationService(queries)
	categoryService := domain.NewCategoryService(queries)
	payeeService := domain.NewPayeeService(queries)
	accountService := domain.NewAccountService(queries, nil)
	trxService := domain.NewTrxService(queries, categoryService, payeeService, accountService) 
	accountService.SetTrxService(trxService)
	goalService := domain.NewGoalService(queries, allocationService)
	return ApiServer{
		budgetService: domain.NewBudgetService(
			queries, 
			trxService,
			allocationService,
			goalService,
			categoryService,
		),
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
	budgets, err := s.budgetService.List(ctx, int(id))
	if err != nil {
		return nil, err
	}
	budgets, err = s.budgetService.List(ctx, int(id))
	if err != nil {
		return nil, err
	}
	budgets, err = s.budgetService.List(ctx, int(id))
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
			Id:   budget.ID(),
			Name: budget.Name(),
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
	if request.Body == nil || request.Body.Name == "" {
		return nil, fmt.Errorf("`name` is required in request body")
	}
	name := request.Body.Name
	// Query the DB
	budget, err := s.budgetService.Create(ctx, int(loginID), name)
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudget201JSONResponse{
		Id:   budget.ID(),
		Name: budget.Name(),
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
	if request.Body == nil || request.Body.Name == "" {
		return nil, fmt.Errorf("`name` is required in request body")
	}
	name := request.Body.Name
	// Query the DB
	budget, err := s.budgetService.Get(ctx, int(loginID), int(budgetID))
	if err != nil {
		return nil, err
	}
	err = budget.Update(ctx, name)
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetId200JSONResponse{
		Id:   budget.ID(),
		Name: budget.Name(),
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
	budget, err := s.budgetService.Get(ctx, int(loginID), int(budgetID))
	if err != nil {
		return nil, err
	}
	err = budget.Delete(ctx)
	if err != nil {
		return nil, err
	}
	// Send the response
	return DeleteBudgetBudgetId204Response{}, nil
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
	budget, err := s.budgetService.Get(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	summary, err := budget.GetMonthSummary(ctx, month)
	if err != nil {
		return nil, err
	}
	categories, err := budget.GetMonthCategories(ctx, month)
	if err != nil {
		return nil, err
	}
	// Send response
	respCategories := make([]BudgetMonthCategory, 0)
	for _, category := range categories {
		var categoryGroupName *string = nil
		if category.Group != nil {
			groupName := category.Group.Name()
			categoryGroupName = &groupName
		}
		var goal *BudgetMonthGoal = nil
		if g := category.Goal; g != nil {
			var endMonth *string = nil
			if g.EndDate() != nil {
				endMonthString := g.EndDate().Format(time.RFC3339)
				endMonth = &endMonthString
			}
			values, err := g.GoalValues(ctx, int(loginID), month)
			if err != nil {
				return nil, err
			}
			goal = &BudgetMonthGoal{
				Type: BudgetMonthGoalType(g.Type()),
				Allocated:      int(values.AllocatedToDate),
				Amount:         int(g.Amount()),
				AmountForMonth: int(values.NeededForMonth),
				Gap:            int(values.Gap),
				StartMonth:     g.StartDate().Format(time.RFC3339),
				EndMonth:       endMonth,
			}
		}
		respCategories = append(respCategories, BudgetMonthCategory{
			CategoryId:        int(category.ID),
			CategoryName:      category.Name,
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
			Allocated:     summary.Allocated,
			Available:     summary.Available,
			Income:        summary.Income,
			ReadyToAssign: summary.ReadyToAssign,
			Spent:         summary.Spent,
			Uncategorized: summary.Uncategorized,
		},
		Categories: respCategories,
	}, nil
}

