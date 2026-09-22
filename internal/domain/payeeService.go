package domain

import (
	"context"
	"database/sql"
	"errors"

	"samuellando.com/YNAFB/data"
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
