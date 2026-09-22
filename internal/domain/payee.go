package domain

import (
	"context"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type Payee struct {
	service *PayeeService
	row data.Payee
}

func (s *PayeeService) fromRow(ctx context.Context, row data.Payee) *Payee {
	if cached, ok := cache.Get[*Payee](ctx, row.ID); ok {
		return cached
	}
	payee := &Payee{
		service: s,
		row: row,
	}
	cache.Store(ctx, row.ID, payee)
	return payee
}

func (p *Payee) ID() int {
	return int(p.row.ID)
}

func (p *Payee) Name() string {
	return p.row.Name
}

// Update renames the payee. loginID is passed explicitly because,
// unlike Account, Payee holds no budget reference to derive it from.
func (p *Payee) Update(ctx context.Context, loginID int, name string) error {
	defer cache.InvalidateResults(ctx)
	row, err := p.service.repo.UpdatePayee(ctx, data.UpdatePayeeParams{
		Name:     name,
		ID:       p.row.ID,
		BudgetID: p.row.BudgetID,
		LoginID:  int64(loginID),
	})
	if err != nil {
		return err
	}
	p.row = row
	return nil
}

// Delete removes the payee. loginID is passed explicitly, see Update.
func (p *Payee) Delete(ctx context.Context, loginID int) error {
	defer cache.InvalidateResults(ctx)
	return p.service.repo.DeletePayee(ctx, data.DeletePayeeParams{
		ID:       p.row.ID,
		BudgetID: p.row.BudgetID,
		LoginID:  int64(loginID),
	})
}
