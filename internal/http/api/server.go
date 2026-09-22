package api

import (
	"database/sql"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db/querycount"
	"samuellando.com/YNAFB/internal/domain"
)

type ApiServer struct {
	budgetService *domain.BudgetService
	accountService *domain.AccountService
	queries       *data.Queries
	db            *sql.DB
}

var _ StrictServerInterface = (*ApiServer)(nil)

func NewServer(db *sql.DB) ApiServer {
	qdb := querycount.New(db)
	queries := data.New(qdb)
	allocationService := domain.NewAllocationService(queries)
	categoryService := domain.NewCategoryService(queries)
	payeeService := domain.NewPayeeService(queries)
	accountService := domain.NewAccountService(queries, nil, nil, payeeService)
	trxService := domain.NewTrxService(queries, categoryService, payeeService, accountService, nil)
	accountService.SetTrxService(trxService)
	goalService := domain.NewGoalService(queries, allocationService)
	budgetService := domain.NewBudgetService(
		queries,
		trxService,
		allocationService,
		goalService,
		categoryService,
	)
	accountService.SetBudgetService(budgetService)
	return ApiServer{
		budgetService: budgetService,
		accountService: accountService,
		queries:       data.New(qdb),
		db:            db,
	}
}

