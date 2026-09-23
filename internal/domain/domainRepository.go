package domain

import (
	"context"
	"samuellando.com/YNAFB/internal/data"
)

type DomainRepository interface {
	// Account Operations
	ListAccounts(context.Context, data.ListAccountsParams) ([]data.ListAccountsRow, error)
	GetAccount(context.Context, data.GetAccountParams) (data.GetAccountRow, error)
	CreateAccount(context.Context, data.CreateAccountParams) (data.CreateAccountRow, error)
	DeleteAccount(context.Context, data.DeleteAccountParams) error
	UpdateAccount(context.Context, data.UpdateAccountParams) (data.Account, error)
	ReconcileAccountTransactions(context.Context, data.ReconcileAccountTransactionsParams) (int64, error)
	// Allocation operations
	ListAllocations(context.Context, data.ListAllocationsParams) ([]data.ListAllocationsRow, error)
	SetAllocation(context.Context, data.SetAllocationParams) (data.Allocation, error)
	// Budget Operations
	CreateBudget(context.Context, data.CreateBudgetParams) (data.Budget, error)
	GetBudget(context.Context, data.GetBudgetParams) (data.Budget, error)
	UpdateBudget(context.Context, data.UpdateBudgetParams) (data.Budget, error)
	DeleteBudget(context.Context, data.DeleteBudgetParams) error
	ListBudgets(context.Context, data.ListBudgetsParams) ([]data.Budget, error)
	// Category operations
	ListCategories(context.Context, data.ListCategoriesParams) ([]data.ListCategoriesRow, error)
	GetCategory(context.Context, data.GetCategoryParams) (data.GetCategoryRow, error)
	CreateCategory(context.Context, data.CreateCategoryParams) (data.Category, error)
	UpdateCategory(context.Context, data.UpdateCategoryParams) (data.Category, error)
	DeleteCategory(context.Context, data.DeleteCategoryParams) error
	// CategoryGroups operations
	ListCategoryGroups(context.Context, data.ListCategoryGroupsParams) ([]data.CategoryGroup, error)
	GetCategoryGroup(context.Context, data.GetCategoryGroupParams) (data.CategoryGroup, error)
	CreateCategoryGroup(context.Context, data.CreateCategoryGroupParams) (int64, error)
	UpdateCategoryGroup(context.Context, data.UpdateCategoryGroupParams) (data.CategoryGroup, error)
	DeleteCategoryGroup(context.Context, data.DeleteCategoryGroupParams) error
	// Goals operations
	ListGoals(context.Context, data.ListGoalsParams) ([]data.ListGoalsRow, error)
	GetGoalByCategory(context.Context, data.GetGoalByCategoryParams) (data.GetGoalByCategoryRow, error)
	CreateGoal(context.Context, data.CreateGoalParams) (data.Goal, error)
	UpdateGoal(context.Context, data.UpdateGoalParams) (data.Goal, error)
	DeleteGoal(context.Context, data.DeleteGoalParams) error
	// Payee operations
	ListPayees(context.Context, data.ListPayeesParams) ([]data.Payee, error)
	GetPayee(context.Context, data.GetPayeeParams) (data.Payee, error)
	GetPayeeByName(context.Context, data.GetPayeeByNameParams) (data.Payee, error)
	CreatePayee(context.Context, data.CreatePayeeParams) (data.Payee, error)
	UpdatePayee(context.Context, data.UpdatePayeeParams) (data.Payee, error)
	DeletePayee(context.Context, data.DeletePayeeParams) error
	// Payee default line operations
	ListPayeeDefaultLinesByPayee(context.Context, data.ListPayeeDefaultLinesByPayeeParams) ([]data.ListPayeeDefaultLinesByPayeeRow, error)
	GetPayeeDefaultLine(context.Context, data.GetPayeeDefaultLineParams) (data.GetPayeeDefaultLineRow, error)
	CreatePayeeDefaultLine(context.Context, data.CreatePayeeDefaultLineParams) (data.PayeeDefaultLine, error)
	UpdatePayeeDefaultLine(context.Context, data.UpdatePayeeDefaultLineParams) (data.PayeeDefaultLine, error)
	DeletePayeeDefaultLine(context.Context, data.DeletePayeeDefaultLineParams) error
	// Trx Operations
	ListTrxsAndLines(context.Context, data.ListTrxsAndLinesParams) ([]data.ListTrxsAndLinesRow, error)
	GetTrxAndLines(context.Context, data.GetTrxAndLinesParams) ([]data.GetTrxAndLinesRow, error)
	CreateTrx(context.Context, data.CreateTrxParams) (data.Trx, error)
	UpdateTrx(context.Context, data.UpdateTrxParams) (data.Trx, error)
	DeleteTrx(context.Context, data.DeleteTrxParams) error
	// TrxLine Operations
	GetTrxLine(context.Context, data.GetTrxLineParams) (data.GetTrxLineRow, error)
	CreateTrxLine(context.Context, data.CreateTrxLineParams) (data.TrxLine, error)
	UpdateTrxLine(context.Context, data.UpdateTrxLineParams) (data.TrxLine, error)
	DeleteTrxLine(context.Context, data.DeleteTrxLineParams) error
}
