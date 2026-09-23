package domain

import (
	"context"
	"database/sql"
	"fmt"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/db/types"
)

type TrxLine struct {
	row                data.TrxLine
	trx                *Trx
	category           *Category
	destinationAccount *Account
}

// Load a trx line from a row. If the row has no line return nil
func trxLineFromRow(ctx context.Context, row data.ListTrxsAndLinesRow, trx *Trx) *TrxLine {
	if !row.LineID.Valid {
		return nil
	}
	line, _ := cache.Get(ctx, row.LineID.Int64, func() (*TrxLine, error) {
		line := &TrxLine{
			row: data.TrxLine{
				ID:            row.LineID.Int64,
				BudgetID:      int64(trx.account.budget.ID()),
				TrxID:         row.Trx.ID,
				DestAccountID: row.DestAccountID,
				CategoryID:    row.CategoryID,
				Income:        row.LineIncome.Bool,
				Inflow:        row.LineInflow.Int64,
				Outflow:       row.LineOutflow.Int64,
			},
			trx: trx,
		}

		if row.CategoryID.Valid {
			var group *CategoryGroup
			if row.CategoryGroupID.Valid {
				group = categoryGroupFromRow(ctx, data.CategoryGroup{
					ID:       row.CategoryGroupID.Int64,
					BudgetID: int64(trx.account.budget.ID()),
					Name:     row.CategoryGroupName.String,
				}, trx.account.budget)
			}
			line.category = categoryFromRow(ctx, data.Category{
				ID:              row.CategoryID.Int64,
				BudgetID:        int64(trx.account.budget.ID()),
				Name:            row.CategoryName.String,
				CategoryGroupID: row.CategoryGroupID,
			}, trx.account.budget, group)
		}
		if row.DestAccountID.Valid {
			line.destinationAccount = accountFromRow(ctx, data.Account{
				ID:       row.DestAccountID.Int64,
				BudgetID: int64(trx.account.budget.ID()),
				Name:     row.DestAccountName.String,
			}, trx.account.budget)
		}
		return line, nil
	})
	return line
}

// Create and add a line to a transaction
func (t *Trx) AddLine(ctx context.Context, destAccount *Account, category *Category, income bool, outflow, inflow int) (*TrxLine, error) {
	defer cache.InvalidateResults(ctx)
	destAccountID := sql.NullInt64{}
	destAccountName := sql.NullString{}
	if destAccount != nil {
		destAccountID = sql.NullInt64{Valid: true, Int64: int64(destAccount.ID())}
		destAccountName = sql.NullString{Valid: true, String: destAccount.Name()}
	}
	categoryID := sql.NullInt64{}
	categoryName := sql.NullString{}
	categoryGroupID := sql.NullInt64{}
	categoryGroupName := sql.NullString{}
	if category != nil {
		categoryID = sql.NullInt64{Valid: true, Int64: int64(category.ID())}
		categoryName = sql.NullString{Valid: true, String: category.Name()}
		if category.group != nil {
			categoryGroupID = sql.NullInt64{Valid: true, Int64: int64(category.group.ID())}
			categoryGroupName = sql.NullString{Valid: true, String: category.group.Name()}
		}
	}
	row, err := t.account.budget.service.repo.CreateTrxLine(ctx, data.CreateTrxLineParams{
		TrxID:         int64(t.ID()),
		DestAccountID: destAccountID,
		CategoryID:    categoryID,
		Income:        income,
		Outflow:       int64(outflow),
		Inflow:        int64(inflow),
		BudgetID:      int64(t.account.budget.ID()),
		LoginID:       int64(t.account.budget.LoginID()),
	})
	if err != nil {
		return nil, err
	}
	line := trxLineFromRow(ctx, data.ListTrxsAndLinesRow{
		Trx: data.Trx{
			ID:           int64(t.ID()),
			BudgetID:     int64(t.account.budget.ID()),
			AccountID:    int64(t.account.ID()),
			PayeeID:      int64(t.payee.ID()),
			Date:         types.UnixTime{Time: t.Date()},
			TotalOutflow: int64(t.TotalOutflow()),
			TotalInflow:  int64(t.TotalInflow()),
			Note:         t.Note(),
		},
		AccountName:       t.account.Name(),
		PayeeName:         t.payee.Name(),
		LineID:            sql.NullInt64{Valid: true, Int64: row.ID},
		LineIncome:        sql.NullBool{Valid: true, Bool: row.Income},
		LineInflow:        sql.NullInt64{Valid: true, Int64: row.Inflow},
		LineOutflow:       sql.NullInt64{Valid: true, Int64: row.Outflow},
		CategoryID:        categoryID,
		CategoryName:      categoryName,
		CategoryGroupID:   categoryGroupID,
		CategoryGroupName: categoryGroupName,
		DestAccountID:     destAccountID,
		DestAccountName:   destAccountName,
		Reconciled:        t.Reconciled(),
	}, t)
	t.lines = append(t.lines, line)
	return line, nil
}

// Get a transaction line
func (t *Trx) GetLine(ctx context.Context, id int) (*TrxLine, error) {
	return cache.Get(ctx, int64(id), func() (*TrxLine, error) {
		row, err := t.account.budget.service.repo.GetTrxLine(ctx, data.GetTrxLineParams{
			LoginID:  int64(t.account.budget.LoginID()),
			BudgetID: int64(t.account.budget.ID()),
			TrxID:    int64(t.ID()),
			ID:       int64(id),
		})
		if err != nil {
			return nil, err
		}
		return trxLineFromRow(ctx, data.ListTrxsAndLinesRow{
			Trx: data.Trx{
				ID:           int64(t.ID()),
				BudgetID:     int64(t.account.budget.ID()),
				AccountID:    int64(t.account.ID()),
				PayeeID:      int64(t.payee.ID()),
				Date:         types.UnixTime{Time: t.Date()},
				TotalOutflow: int64(t.TotalOutflow()),
				TotalInflow:  int64(t.TotalInflow()),
				Note:         t.Note(),
			},
			AccountName:       t.account.Name(),
			PayeeName:         t.payee.Name(),
			LineID:            sql.NullInt64{Valid: true, Int64: row.TrxLine.ID},
			LineIncome:        sql.NullBool{Valid: true, Bool: row.TrxLine.Income},
			LineInflow:        sql.NullInt64{Valid: true, Int64: row.TrxLine.Inflow},
			LineOutflow:       sql.NullInt64{Valid: true, Int64: row.TrxLine.Outflow},
			CategoryID:        row.TrxLine.CategoryID,
			CategoryName:      row.CategoryName,
			CategoryGroupID:   row.CategoryGroupID,
			CategoryGroupName: row.CategoryGroupName,
			DestAccountID:     row.TrxLine.DestAccountID,
			DestAccountName:   row.DestAccountName,
			Reconciled:        t.Reconciled(),
		}, t), nil
	})
}

// Get the transaction line id
func (l *TrxLine) ID() int {
	return int(l.row.ID)
}

// Whether or not the transaction line is income
func (l *TrxLine) IsIncome() bool {
	return l.row.Income
}

// The transaction line category, returns an error if its not a categorization
func (l *TrxLine) Category() (*Category, error) {
	if l.row.CategoryID.Valid {
		return l.category, nil
	}
	return nil, fmt.Errorf("Transaction line is not a categorization")
}

// The transaction line destination account, returns an error if its not a transfer
func (l *TrxLine) DestinationAccount() (*Account, error) {
	if l.row.DestAccountID.Valid {
		return l.destinationAccount, nil
	}
	return nil, fmt.Errorf("Transaction line is not a transfer")
}

// The inflow on this line
func (l *TrxLine) Inflow() int {
	return int(l.row.Inflow)
}

// The outflow on this line
func (l *TrxLine) Outflow() int {
	return int(l.row.Outflow)
}

// The Lines source transaction
func (l *TrxLine) Trx() *Trx {
	return l.trx
}

func (l *TrxLine) Update(ctx context.Context, destAccount *Account, category *Category, income bool, outflow, inflow int) error {
	defer cache.InvalidateResults(ctx)
	destAccountID := sql.NullInt64{}
	if destAccount != nil {
		destAccountID = sql.NullInt64{Valid: true, Int64: int64(destAccount.ID())}
	}
	categoryID := sql.NullInt64{}
	if category != nil {
		categoryID = sql.NullInt64{Valid: true, Int64: int64(category.ID())}
	}
	row, err := l.trx.account.budget.service.repo.UpdateTrxLine(ctx, data.UpdateTrxLineParams{
		ID:            int64(l.ID()),
		TrxID:         int64(l.trx.ID()),
		BudgetID:      int64(l.trx.account.budget.ID()),
		LoginID:       int64(l.trx.account.budget.LoginID()),
		DestAccountID: destAccountID,
		CategoryID:    categoryID,
		Income:        income,
		Outflow:       int64(outflow),
		Inflow:        int64(inflow),
	})
	if err != nil {
		return err
	}
	l.row = row
	l.destinationAccount = destAccount
	l.category = category
	return nil
}

func (l *TrxLine) Delete(ctx context.Context) error {
	defer cache.InvalidateResults(ctx)
	defer cache.Delete[*TrxLine](ctx, int64(l.ID()))
	err := l.trx.account.budget.service.repo.DeleteTrxLine(ctx, data.DeleteTrxLineParams{
		ID:       int64(l.ID()),
		BudgetID: int64(l.trx.account.budget.ID()),
		LoginID:  int64(l.trx.account.budget.LoginID()),
	})
	if err != nil {
		return err
	}
	lines := make([]*TrxLine, 0, len(l.trx.lines))
	for _, existing := range l.trx.lines {
		if existing.ID() != l.ID() {
			lines = append(lines, existing)
		}
	}
	l.trx.lines = lines
	return nil
}
