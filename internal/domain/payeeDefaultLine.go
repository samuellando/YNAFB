package domain

import (
	"context"
	"database/sql"
	"fmt"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type PayeeDefaultLine struct {
	row         data.PayeeDefaultLine
	payee       *Payee
	destAccount *Account
	category    *Category
}

// Create a PayeeDefaultLine object from a data row.
func defaultLineFromRow(ctx context.Context, row data.PayeeDefaultLine, payee *Payee, destAccount *Account, category *Category) *PayeeDefaultLine {
	line, _ := cache.Get(ctx, row.ID, func() (*PayeeDefaultLine, error) {
		return &PayeeDefaultLine{
			row:         row,
			payee:       payee,
			destAccount: destAccount,
			category:    category,
		}, nil
	})
	return line
}

// Add a new default line for the payee
func (p *Payee) AddDefaultLine(ctx context.Context, destAccount *Account, category *Category, income bool, percent int) (*PayeeDefaultLine, error) {
	defer cache.InvalidateResults(ctx)
	destAccountID := sql.NullInt64{}
	if destAccount != nil {
		destAccountID.Valid = true
		destAccountID.Int64 = int64(destAccount.ID())
	}
	categoryID := sql.NullInt64{}
	if category != nil {
		categoryID.Valid = true
		categoryID.Int64 = int64(category.ID())
	}
	row, err := p.budget.service.repo.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		PayeeID:       int64(p.ID()),
		DestAccountID: destAccountID,
		CategoryID:    categoryID,
		Income:        income,
		Percent:       int64(percent),
		BudgetID:      int64(p.budget.ID()),
		LoginID:       int64(p.budget.LoginID()),
	})
	if err != nil {
		return nil, err
	}
	line := defaultLineFromRow(ctx, row, p, destAccount, category)
	return line, nil
}

// Get a payee's default line
func (p *Payee) GetDefaultLine(ctx context.Context, id int) (*PayeeDefaultLine, error) {
	return cache.Get(ctx, int64(id), func() (*PayeeDefaultLine, error) {
		row, err := p.budget.service.repo.GetPayeeDefaultLine(ctx, data.GetPayeeDefaultLineParams{
			LoginID:  int64(p.budget.LoginID()),
			BudgetID: int64(p.budget.ID()),
			ID:       int64(id),
		})
		if err != nil {
			return nil, err
		}
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
				ID:       row.CategoryID.Int64,
				BudgetID: row.BudgetID,
				Name:     row.CategoryName.String,
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
		return defaultLineFromRow(ctx, data.PayeeDefaultLine{
			ID: row.ID,
			BudgetID: row.BudgetID,
			PayeeID: row.PayeeID,
			DestAccountID: row.DestAccountID,
			CategoryID: row.CategoryID,
			Income: row.Income,
			Percent: row.Percent,
		}, p, destAccount, category), nil
	})
}

// Get the line ID
func (l *PayeeDefaultLine) ID() int {
	return int(l.row.ID)
}

func (l *PayeeDefaultLine) Income() bool {
	return l.row.Income
}

func (l *PayeeDefaultLine) Percent() int {
	return int(l.row.Percent)
}

// Get the payee of the default line
func (l *PayeeDefaultLine) Payee() *Payee {
	return l.payee
}

// Get the destination account of the default line, returns an error if none
func (l *PayeeDefaultLine) DestAccount() (*Account, error) {
	if l.destAccount == nil {
		return nil, fmt.Errorf("Default line is not a transfer")
	}
	return l.destAccount, nil
}

// Get the category of the default line, returns an error if none
func (l *PayeeDefaultLine) Category() (*Category, error) {
	if l.category == nil {
		return nil, fmt.Errorf("Default line is not a categorization")
	}
	return l.category, nil
}


func (l *PayeeDefaultLine) Update(ctx context.Context, destAccountID, categoryID *int, income bool, percent int) error {
	defer cache.InvalidateResults(ctx)
	row, err := l.payee.budget.service.repo.UpdatePayeeDefaultLine(ctx, data.UpdatePayeeDefaultLineParams{
		PayeeID:       int64(l.payee.ID()),
		DestAccountID: nullInt64FromInt(destAccountID),
		CategoryID:    nullInt64FromInt(categoryID),
		Income:        income,
		Percent:       int64(percent),
		ID:            int64(l.ID()),
		BudgetID:      int64(l.payee.budget.ID()),
		LoginID:       int64(l.payee.budget.LoginID()),
	})
	if err != nil {
		return err
	}
	l.row = row
	return nil
}

func (l *PayeeDefaultLine) Delete(ctx context.Context) error {
	defer cache.InvalidateResults(ctx)
	defer cache.Delete[*PayeeDefaultLine](ctx, int64(l.ID()))
	return l.payee.budget.service.repo.DeletePayeeDefaultLine(ctx, data.DeletePayeeDefaultLineParams{
		ID:       int64(l.ID()),
		BudgetID: int64(l.payee.budget.ID()),
		LoginID:  int64(l.payee.budget.LoginID()),
	})
}

