package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type PayeeService struct {
	repo PayeeRepository
}

func NewPayeeService(repo PayeeRepository) *PayeeService {
	return &PayeeService{repo: repo}
}

// WithRepo returns a copy of the service bound to the given repo,
// for use inside a transaction.
func (s *PayeeService) WithRepo(repo PayeeRepository) *PayeeService {
	cp := *s
	cp.repo = repo
	return &cp
}

func (s *PayeeService) GetOrCreate(ctx context.Context, loginID, budgetID int, name string) (*Payee, error) {
	payee, err := s.repo.GetPayeeByName(ctx, data.GetPayeeByNameParams{
		BudgetID: int64(budgetID),
		LoginID:  int64(loginID),
		Name:     name,
	})
	if err == nil {
		return s.fromRow(ctx, payee), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	created, err := s.repo.CreatePayee(ctx, data.CreatePayeeParams{
		BudgetID: int64(budgetID),
		LoginID:  int64(loginID),
		Name:     name,
	})
	if err != nil {
		return nil, err
	}
	return s.fromRow(ctx, created), nil
}

func (s *PayeeService) List(ctx context.Context, loginID, budgetID int) ([]*Payee, error) {
	return cache.Result(ctx, fmt.Sprintf("payeeServiceList-%d-%d", loginID, budgetID), func() ([]*Payee, error) {
		return s.list(ctx, loginID, budgetID)
	})
}

func (s *PayeeService) list(ctx context.Context, loginID, budgetID int) ([]*Payee, error) {
	rows, err := s.repo.ListPayees(ctx, data.ListPayeesParams{
		LoginID:  int64(loginID),
		BudgetID: int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	payees := make([]*Payee, len(rows))
	for i, row := range rows {
		payees[i] = s.fromRow(ctx, row)
	}
	return payees, nil
}

func (s *PayeeService) Create(ctx context.Context, loginID, budgetID int, name string) (*Payee, error) {
	defer cache.InvalidateResults(ctx)
	row, err := s.repo.CreatePayee(ctx, data.CreatePayeeParams{
		Name:     name,
		BudgetID: int64(budgetID),
		LoginID:  int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	payee := s.fromRow(ctx, row)
	return payee, nil
}

func (s *PayeeService) Get(ctx context.Context, loginID, budgetID, payeeID int) (*Payee, error) {
	if v, ok := cache.Get[*Payee](ctx, int64(payeeID)); ok {
		return v, nil
	}
	row, err := s.repo.GetPayee(ctx, data.GetPayeeParams{
		LoginID:  int64(loginID),
		BudgetID: int64(budgetID),
		ID:       int64(payeeID),
	})
	if err != nil {
		return nil, err
	}
	payee := s.fromRow(ctx, row)
	return payee, nil
}

func (s *PayeeService) ListDefaultLinesByPayee(ctx context.Context, loginID, budgetID, payeeID int) ([]data.ListPayeeDefaultLinesByPayeeRow, error) {
	return cache.Result(ctx, fmt.Sprintf("payeeServiceListDefaultLinesByPayee-%d-%d-%d", loginID, budgetID, payeeID), func() ([]data.ListPayeeDefaultLinesByPayeeRow, error) {
		return s.repo.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{
			LoginID:  int64(loginID),
			BudgetID: int64(budgetID),
			PayeeID:  int64(payeeID),
		})
	})
}

func (s *PayeeService) CreateDefaultLine(ctx context.Context, loginID, budgetID, payeeID int, destAccountID, categoryID *int, income bool, percent int) (*PayeeDefaultLine, error) {
	defer cache.InvalidateResults(ctx)
	row, err := s.repo.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		PayeeID:       int64(payeeID),
		DestAccountID: nullInt64FromInt(destAccountID),
		CategoryID:    nullInt64FromInt(categoryID),
		Income:        income,
		Percent:       int64(percent),
		BudgetID:      int64(budgetID),
		LoginID:       int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	line := s.defaultLineFromRow(ctx, row)
	cache.Store(ctx, line.row.ID, line)
	return line, nil
}

func (s *PayeeService) GetDefaultLine(ctx context.Context, loginID, budgetID, id int) (*PayeeDefaultLine, error) {
	if v, ok := cache.Get[*PayeeDefaultLine](ctx, int64(id)); ok {
		return v, nil
	}
	row, err := s.repo.GetPayeeDefaultLine(ctx, data.GetPayeeDefaultLineParams{
		LoginID:  int64(loginID),
		BudgetID: int64(budgetID),
		ID:       int64(id),
	})
	if err != nil {
		return nil, err
	}
	line := s.defaultLineFromRow(ctx, row)
	cache.Store(ctx, line.row.ID, line)
	return line, nil
}
