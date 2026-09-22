package domain

import (
	"context"
	"fmt"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type AccountService struct {
	repo          AccountRepository
	trxService    *TrxService
	budgetService *BudgetService
	payeeService  *PayeeService
}

func NewAccountService(repo AccountRepository, trxService *TrxService, budgetService *BudgetService, payeeService *PayeeService) *AccountService {
	return &AccountService{
		repo:          repo,
		trxService:    trxService,
		budgetService: budgetService,
		payeeService:  payeeService,
	}
}

func (s *AccountService) SetTrxService(trxService *TrxService) {
	s.trxService = trxService
}

func (s *AccountService) SetBudgetService(budgetService *BudgetService) {
	s.budgetService = budgetService
}

func (s *AccountService) SetPayeeService(payeeService *PayeeService) {
	s.payeeService = payeeService
}

// AccountRepos is the set of repos an AccountService needs. *data.Queries
// satisfies it, including tx-scoped copies from Queries.WithTx.
type AccountRepos interface {
	AccountRepository
	TrxRepository
	PayeeRepository
}

// WithRepo returns a copy of the service (including its payee and trx
// services) bound to the given repos, for use inside a transaction.
func (s *AccountService) WithRepo(repos AccountRepos) *AccountService {
	cp := *s
	cp.repo = repos
	if s.payeeService != nil {
		cp.payeeService = s.payeeService.WithRepo(repos)
	}
	if s.trxService != nil {
		cp.trxService = s.trxService.WithRepo(repos)
	}
	return &cp
}

func (s *AccountService) List(ctx context.Context, loginID, budgetID int) ([]*Account, error) {
	return cache.Result(ctx, fmt.Sprintf("accountServiceList-%d-%d", loginID, budgetID), func() ([]*Account, error) {
		return s.list(ctx, loginID, budgetID)
	})
}

func (s *AccountService) list(ctx context.Context, loginID, budgetID int) ([]*Account, error) {
	rows, err := s.repo.ListAccounts(ctx, data.ListAccountsParams{
		LoginID:  int64(loginID),
		BudgetID: int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	accounts := make([]*Account, len(rows))
	for i, row := range rows {
		accounts[i] = s.fromRow(ctx, data.Account{
			ID:       row.ID,
			BudgetID: row.BudgetID,
			Name:     row.Name,
		}, s.budgetService.fromRow(ctx, data.Budget{
			ID:      row.BudgetID,
			LoginID: row.LoginID,
			Name:    row.BudgetName,
		}))
		cache.Store(ctx, accounts[i].row.ID, accounts[i])
	}
	return accounts, nil
}

func (s *AccountService) Create(ctx context.Context, loginID, budgetID int, name string) (*Account, error) {
	defer cache.InvalidateResults(ctx)
	row, err := s.repo.CreateAccount(ctx, data.CreateAccountParams{
		Name:     name,
		LoginID:  int64(loginID),
		BudgetID: int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	account := s.fromRow(ctx, data.Account{
		ID:       row.ID,
		BudgetID: row.BudgetID,
		Name:     row.Name,
	}, s.budgetService.fromRow(ctx, data.Budget{
		ID:      row.BudgetID,
		LoginID: row.LoginID,
		Name:    row.BudgetName,
	}))
	cache.Store(ctx, account.row.ID, account)
	return account, nil
}

func (s *AccountService) Get(ctx context.Context, loginID, budgetID, accountID int) (*Account, error) {
	if v, ok := cache.Get[*Account](ctx, int64(accountID)); ok {
		return v, nil
	}
	row, err := s.repo.GetAccount(ctx, data.GetAccountParams{
		LoginID:  int64(loginID),
		BudgetID: int64(budgetID),
		ID:       int64(accountID),
	})
	if err != nil {
		return nil, err
	}
	account := s.fromRow(ctx, data.Account{
		ID:       row.ID,
		BudgetID: row.BudgetID,
		Name:     row.Name,
	}, s.budgetService.fromRow(ctx, data.Budget{
		ID:      row.BudgetID,
		LoginID: row.LoginID,
		Name:    row.BudgetName,
	}))
	cache.Store(ctx, account.row.ID, account)
	return account, nil
}
