package domain

import (
	"context"
	"samuellando.com/YNAFB/data"
)

type BudgetRepository interface {
	// Budget Operations
	CreateBudget(context.Context, data.CreateBudgetParams) (data.Budget, error)
	GetBudget(context.Context, data.GetBudgetParams) (data.Budget, error)
	UpdateBudget(context.Context, data.UpdateBudgetParams) (data.Budget, error)
	DeleteBudget(context.Context, data.DeleteBudgetParams) (error)
	ListBudgets(context.Context, data.ListBudgetsParams) ([]data.Budget, error)
}
