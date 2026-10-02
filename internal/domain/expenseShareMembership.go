package domain

import (
	"context"
	"fmt"

	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/data"
)

type ExpenseShareMembership struct {
	row    data.BudgetExpenseShare
	budget *Budget
	share  *ExpenseShare
}

type ShareMember struct {
	Budget      *Budget
	DisplayName string
}

type MemberBalance struct {
	Budget      *Budget
	DisplayName string
	Balance     int
}

func expenseShareMembershipFromRow(ctx context.Context, row data.BudgetExpenseShare, budget *Budget, share *ExpenseShare) *ExpenseShareMembership {
	membership, _ := cache.Get(ctx, row.ID, func() (*ExpenseShareMembership, error) {
		return &ExpenseShareMembership{
			row:    row,
			budget: budget,
			share:  share,
		}, nil
	})
	return membership
}

func (b *Budget) ListExpenseShares(ctx context.Context) ([]*ExpenseShareMembership, error) {
	return nil, fmt.Errorf("expense share not implemented")
}

func (b *Budget) CreateExpenseShare(ctx context.Context, name, displayName string, defaultName *string) (*ExpenseShareMembership, error) {
	return nil, fmt.Errorf("expense share not implemented")
}

func (b *Budget) JoinExpenseShare(ctx context.Context, code, name, displayName string) (*ExpenseShareMembership, error) {
	return nil, fmt.Errorf("expense share not implemented")
}

func (b *Budget) GetExpenseShare(ctx context.Context, shareID int) (*ExpenseShareMembership, error) {
	return nil, fmt.Errorf("expense share not implemented")
}

func (m *ExpenseShareMembership) ID() int {
	return int(m.row.ID)
}

func (m *ExpenseShareMembership) Name() string {
	return m.row.Name
}

func (m *ExpenseShareMembership) DisplayName() string {
	return m.row.DisplayName
}

func (m *ExpenseShareMembership) Budget() *Budget {
	return m.budget
}

func (m *ExpenseShareMembership) Share() *ExpenseShare {
	return m.share
}

func (m *ExpenseShareMembership) Members(ctx context.Context) ([]*ShareMember, error) {
	return nil, fmt.Errorf("expense share not implemented")
}

func (m *ExpenseShareMembership) Balances(ctx context.Context) (int, []MemberBalance, error) {
	return 0, nil, fmt.Errorf("expense share not implemented")
}

func (m *ExpenseShareMembership) Update(ctx context.Context, name, displayName *string) error {
	return fmt.Errorf("expense share not implemented")
}

func (m *ExpenseShareMembership) Leave(ctx context.Context) error {
	return fmt.Errorf("expense share not implemented")
}

func (m *ExpenseShareMembership) ListTransactions(ctx context.Context) ([]*ExpenseShareTransaction, error) {
	return nil, fmt.Errorf("expense share not implemented")
}

func (m *ExpenseShareMembership) GetTransaction(ctx context.Context, trxID int) (*ExpenseShareTransaction, error) {
	return nil, fmt.Errorf("expense share not implemented")
}
