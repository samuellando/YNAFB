package domain

import (
	"context"
	"database/sql"
	"fmt"

	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/data"
)

type PayeeDefaultLine struct {
	row         data.PayeeDefaultLine
	payee       *Payee
	destAccount *Account
	category    *Category
	// Expense share qualifier (present on split and settlement lines only).
	// Split/dest member budgets resolve via share-scoped member queries
	// (queries slice) and stay nil until then.
	share       *ExpenseShare
	splitBudget *Budget
	destBudget  *Budget
}

// Create a PayeeDefaultLine object from a data row. Share-typed relations
// arrive as objects; call sites pass nil until the share-scoped member
// queries land.
func defaultLineFromRow(ctx context.Context, row data.PayeeDefaultLine, payee *Payee, destAccount *Account, category *Category, share *ExpenseShare, splitBudget, destBudget *Budget) *PayeeDefaultLine {
	line, _ := cache.Get(ctx, row.ID, func() (*PayeeDefaultLine, error) {
		return &PayeeDefaultLine{
			row:         row,
			payee:       payee,
			destAccount: destAccount,
			category:    category,
			share:       share,
			splitBudget: splitBudget,
			destBudget:  destBudget,
		}, nil
	})
	return line
}

// Add a new default line for the payee
func (p *Payee) AddDefaultLine(ctx context.Context, destAccount *Account, category *Category, income bool, expenseShare *ExpenseShare, splitBudget, destBudget *Budget, percent int) (*PayeeDefaultLine, error) {
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
	expenseShareID := sql.NullInt64{}
	if expenseShare != nil {
		expenseShareID.Valid = true
		expenseShareID.Int64 = int64(expenseShare.ID())
	}
	splitBudgetID := sql.NullInt64{}
	if splitBudget != nil {
		splitBudgetID.Valid = true
		splitBudgetID.Int64 = int64(splitBudget.ID())
	}
	destBudgetID := sql.NullInt64{}
	if destBudget != nil {
		destBudgetID.Valid = true
		destBudgetID.Int64 = int64(splitBudget.ID())
	}
	row, err := p.budget.service.repo.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
		PayeeID:        int64(p.ID()),
		DestAccountID:  destAccountID,
		CategoryID:     categoryID,
		ExpenseShareID: expenseShareID,
		SplitBudgetID:  splitBudgetID,
		DestBudgetID:   destBudgetID,
		Income:         income,
		Percent:        int64(percent),
		BudgetID:       int64(p.budget.ID()),
		LoginID:        int64(p.budget.LoginID()),
	})
	if err != nil {
		return nil, err
	}
	line := defaultLineFromRow(ctx, row, p, destAccount, category, expenseShare, splitBudget, destBudget)
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
		var expenseShare *ExpenseShare
		if row.ExpenseShareID.Valid {
			expenseShare = expenseShareFromRow(ctx, data.ExpenseShare{
				ID:          row.ExpenseShareID.Int64,
				DefaultName: row.ExpenseShareDefaultName.String,
			}, p.budget.service)
		}
		var splitBudget *Budget
		if row.SplitBudgetID.Valid {
			splitBudget = budgetFromRow(ctx, data.Budget{
				ID:      row.SplitBudgetID.Int64,
				LoginID: row.SplitBudgetLoginID.Int64,
				Name:    row.SplitBudgetName.String,
			}, p.budget.service)
		}
		var destBudget *Budget
		if row.DestBudgetID.Valid {
			splitBudget = budgetFromRow(ctx, data.Budget{
				ID:      row.DestBudgetID.Int64,
				LoginID: row.DestBudgetLoginID.Int64,
				Name:    row.DestBudgetName.String,
			}, p.budget.service)
		}
		return defaultLineFromRow(ctx, data.PayeeDefaultLine{
			ID:             row.ID,
			BudgetID:       row.BudgetID,
			PayeeID:        row.PayeeID,
			DestAccountID:  row.DestAccountID,
			CategoryID:     row.CategoryID,
			ExpenseShareID: row.ExpenseShareID,
			SplitBudgetID:  row.SplitBudgetID,
			DestBudgetID:   row.DestBudgetID,
			Income:         row.Income,
			Percent:        row.Percent,
		}, p, destAccount, category, expenseShare, splitBudget, destBudget), nil
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

// Get the expense share of the default line, returns an error if none
func (l *PayeeDefaultLine) Share() (*ExpenseShare, error) {
	if l.row.ExpenseShareID.Valid {
		return l.share, nil
	}
	return nil, fmt.Errorf("Default line is not a split or settlement line")
}

// Get the split line tagged member budget, returns an error if none
func (l *PayeeDefaultLine) SplitBudget() (*Budget, error) {
	if l.row.SplitBudgetID.Valid {
		return l.splitBudget, nil
	}
	return nil, fmt.Errorf("Default line is not a split line")
}

// Get the settlement line counterparty member budget, returns an error if none
func (l *PayeeDefaultLine) DestBudget() (*Budget, error) {
	if l.row.DestBudgetID.Valid {
		return l.destBudget, nil
	}
	return nil, fmt.Errorf("Default line is not a settlement line")
}

func (l *PayeeDefaultLine) Update(ctx context.Context, destAccountID, categoryID *int, income bool, expenseShareID, splitBudgetID, destBudgetID *int, percent int) error {
	defer cache.InvalidateResults(ctx)
	// TODO: this is should take in the actual domain objects, derive the IDs from them and update the pointers in the pdl
	row, err := l.payee.budget.service.repo.UpdatePayeeDefaultLine(ctx, data.UpdatePayeeDefaultLineParams{
		PayeeID:        int64(l.payee.ID()),
		DestAccountID:  nullInt64FromInt(destAccountID),
		CategoryID:     nullInt64FromInt(categoryID),
		ExpenseShareID: nullInt64FromInt(expenseShareID),
		SplitBudgetID:  nullInt64FromInt(splitBudgetID),
		DestBudgetID:   nullInt64FromInt(destBudgetID),
		Income:         income,
		Percent:        int64(percent),
		ID:             int64(l.ID()),
		BudgetID:       int64(l.payee.budget.ID()),
		LoginID:        int64(l.payee.budget.LoginID()),
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
