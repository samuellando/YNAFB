package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/data"
	"samuellando.com/YNAFB/internal/db/types"
)

type ExpenseShare struct {
	service *DomainService
	row data.ExpenseShare
}

func expenseShareFromRow(ctx context.Context, row data.ExpenseShare, service *DomainService) *ExpenseShare {
	share, _ := cache.Get(ctx, row.ID, func() (*ExpenseShare, error) {
		return &ExpenseShare{
			service: service,
			row: row,
		}, nil
	})
	return share
}

func (s *DomainService) GetExpenseShare(ctx context.Context, id int) (*ExpenseShare, error) {
	return cache.Get(ctx, int64(id), func() (*ExpenseShare, error) {
		row, err := s.repo.GetExpenseShareById(ctx, data.GetExpenseShareByIdParams{
			ID: int64(id),
		})
		if err != nil {
			return nil, err
		}
		return expenseShareFromRow(ctx, row, s), nil
	})
}

func (s *DomainService) createExpenseShare(ctx context.Context, defaultName string) (*ExpenseShare, error) {
	defer cache.InvalidateResults(ctx)
	row, err := s.repo.CreateExpenseShare(ctx, data.CreateExpenseShareParams{
		DefaultName: defaultName,
	})
	if err != nil {
		return nil, err
	}
	return expenseShareFromRow(ctx, row, s), nil
}

func (e *ExpenseShare) delete(ctx context.Context, service *DomainService) error {
	defer cache.InvalidateResults(ctx)
	defer cache.Delete[*ExpenseShare](ctx, int64(e.ID()))
	return service.repo.DeleteExpenseShare(ctx, data.DeleteExpenseShareParams{
		ID: int64(e.ID()),
	})
}

func (e *ExpenseShare) ID() int {
	return int(e.row.ID)
}

func (e *ExpenseShare) DefaultName() string {
	return e.row.DefaultName
}

func (s *ExpenseShare) MintCode(ctx context.Context) (*ExpenseShareCode, error) {
	defer cache.InvalidateResults(ctx)
	row, err := s.service.repo.CreateExpenseShareCode(ctx, data.CreateExpenseShareCodeParams{
		ExpenseShareID: int64(s.ID()),
		Code:           uuid.NewString(),
		Expires:        types.UnixTime{Time: time.Now().Add(7 * 24 * time.Hour)},
	})
	if err != nil {
		return nil, err
	}
	return expenseShareCodeFromRow(ctx, row, s), nil
}

func (m *ExpenseShare) GetMembership(ctx context.Context, budgetID int) (*ExpenseShareMembership, error) {
	return cache.Result(ctx, fmt.Sprintf("shareMemberByBudget-%d-%d", m.ID(), budgetID), func() (*ExpenseShareMembership, error) {
		row, err := m.service.repo.GetShareMemberByBudget(ctx, data.GetShareMemberByBudgetParams{
			ExpenseShareID: int64(m.ID()),
			BudgetID:       int64(budgetID),
		})
		if err != nil {
			return nil, err
		}
		return membershipFromMemberRow(ctx, row.BudgetExpenseShare, row.ExpenseShare, row.Budget, m.service), nil
	})
}
