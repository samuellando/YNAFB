package domain

import (
	"context"
	"fmt"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type Payee struct {
	row    data.Payee
	budget *Budget
}

// Create a Payee object from a data row.
func payeeFromRow(ctx context.Context, row data.Payee, budget *Budget) *Payee {
	payee, _ := cache.Get(ctx, row.ID, func() (*Payee, error) {
		return &Payee{
			row:    row,
			budget: budget,
		}, nil
	})
	return payee
}

// Create a new payee
func (b *Budget) CreatePayee(ctx context.Context, name string) (*Payee, error) {
	defer cache.InvalidateResults(ctx)
	row, err := b.service.repo.CreatePayee(ctx, data.CreatePayeeParams{
		Name:     name,
		BudgetID: int64(b.ID()),
		LoginID:  int64(b.LoginID()),
	})
	if err != nil {
		return nil, err
	}
	payee := payeeFromRow(ctx, row, b)
	return payee, nil
}

// Get an existing payee by ID
func (b *Budget) GetPayee(ctx context.Context, payeeID int) (*Payee, error) {
	return cache.Get(ctx, int64(payeeID), func() (*Payee, error) {
		row, err := b.service.repo.GetPayee(ctx, data.GetPayeeParams{
			LoginID:  int64(b.LoginID()),
			BudgetID: int64(b.ID()),
			ID:       int64(payeeID),
		})
		if err != nil {
			return nil, err
		}
		return payeeFromRow(ctx, row, b), nil
	})
}

// Look up a payee by name, if it does not exist create it
func (b *Budget) getOrCreatePayee(ctx context.Context, name string) (*Payee, error) {
	row, err := b.service.repo.GetPayeeByName(ctx, data.GetPayeeByNameParams{
		BudgetID: int64(b.ID()),
		LoginID:  int64(b.LoginID()),
		Name:     name,
	})
	if err != nil {
		return b.CreatePayee(ctx, name)
	}
	return payeeFromRow(ctx, row, b), nil
}

// List all the payees in the budget
func (b *Budget) ListPayees(ctx context.Context) ([]*Payee, error) {
	return cache.Result(ctx, fmt.Sprintf("payeeServiceList-%d-%d", b.LoginID(), b.ID()), func() ([]*Payee, error) {
		rows, err := b.service.repo.ListPayees(ctx, data.ListPayeesParams{
			LoginID:  int64(b.LoginID()),
			BudgetID: int64(b.ID()),
		})
		if err != nil {
			return nil, err
		}
		payees := make([]*Payee, len(rows))
		for i, row := range rows {
			payees[i] = payeeFromRow(ctx, row, b)
		}
		return payees, nil
	})
}

// Update the existing payee
func (p *Payee) Update(ctx context.Context, name string) error {
	defer cache.InvalidateResults(ctx)
	row, err := p.budget.service.repo.UpdatePayee(ctx, data.UpdatePayeeParams{
		Name:     name,
		ID:       int64(p.ID()),
		BudgetID: int64(p.budget.ID()),
		LoginID:  int64(p.budget.LoginID()),
	})
	if err != nil {
		return err
	}
	p.row = row
	return nil
}

// Delete a payee
func (p *Payee) Delete(ctx context.Context) error {
	defer cache.InvalidateResults(ctx)
	defer cache.Delete[*Payee](ctx, int64(p.ID()))
	return p.budget.service.repo.DeletePayee(ctx, data.DeletePayeeParams{
		ID:       int64(p.ID()),
		BudgetID: int64(p.budget.ID()),
		LoginID:  int64(p.budget.LoginID()),
	})
}

// Get the payee's ID
func (p *Payee) ID() int {
	return int(p.row.ID)
}

// Get the payee's name
func (p *Payee) Name() string {
	return p.row.Name
}

// List the payee's default lines
func (p *Payee) DefaultLines(ctx context.Context) ([]*PayeeDefaultLine, error) {
	return cache.Result(ctx, fmt.Sprintf("payeeServiceListDefaultLinesByPayee-%d-%d-%d", p.budget.LoginID(), p.budget.ID(), p.ID()), func() ([]*PayeeDefaultLine, error) {
		lines := make([]*PayeeDefaultLine, 0)
		rows, err := p.budget.service.repo.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{
			LoginID:  int64(p.budget.LoginID()),
			BudgetID: int64(p.budget.ID()),
			PayeeID:  int64(p.ID()),
		})
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			var categoryGroup *CategoryGroup
			if row.CategoryGroupID.Valid {
				categoryGroup = categoryGroupFromRow(ctx, data.CategoryGroup{
					ID:       row.CategoryGroupID.Int64,
					BudgetID: row.BudgetID,
					Name:     row.CategoryGroupName.String,
				}, p.budget)
			}
			var category *Category
			if row.CategoryID.Valid {
				category = categoryFromRow(ctx, data.Category{
					ID:              row.CategoryID.Int64,
					BudgetID:        row.BudgetID,
					Name:            row.CategoryName.String,
					CategoryGroupID: row.CategoryGroupID,
				}, p.budget, categoryGroup)
			}
			var destAccount *Account
			if row.DestAccountID.Valid {
				destAccount = accountFromRow(ctx, data.Account{
					ID:       row.DestAccountID.Int64,
					BudgetID: row.BudgetID,
					Name:     row.DestAccountName.String,
				}, p.budget)
			}
			lines = append(lines, defaultLineFromRow(ctx, data.PayeeDefaultLine{
				ID:            row.ID,
				BudgetID:      row.BudgetID,
				PayeeID:       row.PayeeID,
				DestAccountID: row.DestAccountID,
				CategoryID:    row.CategoryID,
				Income:        row.Income,
				Percent:       row.Percent,
			}, p, destAccount, category))
		}
		return lines, nil
	})
}
