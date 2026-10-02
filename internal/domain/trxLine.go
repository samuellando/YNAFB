package domain

import (
	"context"
	"database/sql"
	"fmt"

	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/data"
	"samuellando.com/YNAFB/internal/db/types"
)

type TrxLine struct {
	row                data.TrxLine
	trx                *Trx
	category           *Category
	destinationAccount *Account
	share              *ExpenseShare
	splitBudget        *ExpenseShareMembership
	destBudget         *ExpenseShareMembership
}

// Load a trx line from a row.
func trxLineFromRow(ctx context.Context, row data.ListTrxsAndLinesRow, trx *Trx) *TrxLine {
	if !row.LineID.Valid {
		return nil
	}
	line, _ := cache.Get(ctx, row.LineID.Int64, func() (*TrxLine, error) {
		line := &TrxLine{
			row: data.TrxLine{
				ID:                        row.LineID.Int64,
				BudgetID:                  int64(trx.account.budget.ID()),
				TrxID:                     row.Trx.ID,
				DestAccountID:             row.DestAccountID,
				CategoryID:                row.CategoryID,
				Income:                    row.LineIncome.Bool,
				ExpenseShareID:            row.ExpenseShareID,
				SplitBudgetExpenseShareID: row.SplitBudgetExpenseShareID,
				DestBudgetExpenseShareID:  row.DestBudgetExpenseShareID,
				Inflow:                    row.LineInflow.Int64,
				Outflow:                   row.LineOutflow.Int64,
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
		if row.ExpenseShareID.Valid {
			line.share = expenseShareFromRow(ctx, data.ExpenseShare{
				ID:          row.ExpenseShareID.Int64,
				DefaultName: row.ExpenseShareDefaultName.String,
			}, trx.account.budget.service)
		}
		if row.SplitBudgetExpenseShareID.Valid {
			splitBudget := budgetFromRow(ctx, data.Budget{
				ID:      row.SplitBudgetID.Int64,
				Name:    row.SplitBudgetName.String,
				LoginID: row.SplitBudgetLoginID.Int64,
			}, trx.account.budget.service)
			line.splitBudget = expenseShareMembershipFromRow(ctx, data.BudgetExpenseShare{
				ID:          row.SplitBudgetExpenseShareID.Int64,
				Name:        row.SplitBudgetExpenseShareName.String,
				DisplayName: row.SplitBudgetExpenseShareDisplayName.String,
				BudgetID:    row.SplitBudgetID.Int64,
			}, splitBudget, line.share)
		}
		if row.DestBudgetID.Valid {
			destBudget := budgetFromRow(ctx, data.Budget{
				ID:      row.DestBudgetID.Int64,
				Name:    row.DestBudgetName.String,
				LoginID: row.DestBudgetLoginID.Int64,
			}, trx.account.budget.service)
			line.destBudget = expenseShareMembershipFromRow(ctx, data.BudgetExpenseShare{
				ID:          row.DestBudgetExpenseShareID.Int64,
				Name:        row.DestBudgetExpenseShareName.String,
				DisplayName: row.DestBudgetExpenseShareDisplayName.String,
				BudgetID:    row.DestBudgetID.Int64,
			}, destBudget, line.share)
		}
		return line, nil
	})
	return line
}

// Create and add a line to a transaction
func (t *Trx) AddLine(ctx context.Context, destAccount *Account, category *Category, income bool, share *ExpenseShare, splitBudget, destBudget *ExpenseShareMembership, outflow, inflow int) (*TrxLine, error) {
	if t.IsMirror() {
		return nil, fmt.Errorf("Mirror transactions cannot be edited")
	}
	if destAccount != nil && destAccount.ID() == t.account.ID() {
		return nil, fmt.Errorf("Transfer destination must differ from source account")
	}
	for _, existing := range t.lines {
		if _, err := existing.DestBudget(); err == nil {
			return nil, fmt.Errorf("Settlement transactions hold exactly one line")
		}
	}
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
	shareID := sql.NullInt64{}
	shareDefaultName := sql.NullString{}
	if share != nil {
		shareID = sql.NullInt64{Valid: true, Int64: int64(share.ID())}
		shareDefaultName = sql.NullString{Valid: true, String: share.DefaultName()}
	}
	splitBudgetExpenseShareID := sql.NullInt64{}
	splitBudgetID := sql.NullInt64{}
	splitBudgetName := sql.NullString{}
	splitBudgetLoginID := sql.NullInt64{}
	if splitBudget != nil {
		splitBudgetExpenseShareID = sql.NullInt64{Valid: true, Int64: int64(splitBudget.ID())}
		splitBudgetID = sql.NullInt64{Valid: true, Int64: int64(splitBudget.Budget().ID())}
		splitBudgetName = sql.NullString{Valid: true, String: splitBudget.Budget().Name()}
		splitBudgetLoginID = sql.NullInt64{Valid: true, Int64: int64(splitBudget.Budget().LoginID())}
	}
	destBudgetExpenseShareID := sql.NullInt64{}
	destBudgetID := sql.NullInt64{}
	destBudgetName := sql.NullString{}
	destBudgetLoginID := sql.NullInt64{}
	if destBudget != nil {
		destBudgetExpenseShareID = sql.NullInt64{Valid: true, Int64: int64(destBudget.ID())}
		destBudgetID = sql.NullInt64{Valid: true, Int64: int64(destBudget.Budget().ID())}
		destBudgetName = sql.NullString{Valid: true, String: destBudget.Budget().Name()}
		destBudgetLoginID = sql.NullInt64{Valid: true, Int64: int64(destBudget.Budget().LoginID())}
	}
	row, err := t.account.budget.service.repo.CreateTrxLine(ctx, data.CreateTrxLineParams{
		TrxID:                     int64(t.ID()),
		DestAccountID:             destAccountID,
		CategoryID:                categoryID,
		Income:                    income,
		ExpenseShareID:            shareID,
		SplitBudgetExpenseShareID: splitBudgetExpenseShareID,
		DestBudgetExpenseShareID:  destBudgetExpenseShareID,
		Outflow:                   int64(outflow),
		Inflow:                    int64(inflow),
		BudgetID:                  int64(t.account.budget.ID()),
		LoginID:                   int64(t.account.budget.LoginID()),
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
		AccountName:               t.account.Name(),
		PayeeName:                 t.payee.Name(),
		LineID:                    sql.NullInt64{Valid: true, Int64: row.ID},
		LineIncome:                sql.NullBool{Valid: true, Bool: row.Income},
		LineInflow:                sql.NullInt64{Valid: true, Int64: row.Inflow},
		LineOutflow:               sql.NullInt64{Valid: true, Int64: row.Outflow},
		CategoryID:                categoryID,
		CategoryName:              categoryName,
		CategoryGroupID:           categoryGroupID,
		CategoryGroupName:         categoryGroupName,
		DestAccountID:             destAccountID,
		DestAccountName:           destAccountName,
		ExpenseShareID:            shareID,
		ExpenseShareDefaultName:   shareDefaultName,
		SplitBudgetExpenseShareID: splitBudgetExpenseShareID,
		SplitBudgetID:             splitBudgetID,
		SplitBudgetName:           splitBudgetName,
		SplitBudgetLoginID:        splitBudgetLoginID,
		DestBudgetExpenseShareID:  destBudgetExpenseShareID,
		DestBudgetID:              destBudgetID,
		DestBudgetName:            destBudgetName,
		DestBudgetLoginID:         destBudgetLoginID,
		Reconciled:                t.Reconciled(),
	}, t)
	t.lines = append(t.lines, line)
	return line, nil
}

// Get a transaction line
func (t *Trx) GetLine(ctx context.Context, id int) (*TrxLine, error) {
	if t.IsMirror() {
		return nil, fmt.Errorf("Mirror transactions cannot be edited")
	}
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
			AccountName:               t.account.Name(),
			PayeeName:                 t.payee.Name(),
			LineID:                    sql.NullInt64{Valid: true, Int64: row.TrxLine.ID},
			LineIncome:                sql.NullBool{Valid: true, Bool: row.TrxLine.Income},
			LineInflow:                sql.NullInt64{Valid: true, Int64: row.TrxLine.Inflow},
			LineOutflow:               sql.NullInt64{Valid: true, Int64: row.TrxLine.Outflow},
			CategoryID:                row.TrxLine.CategoryID,
			CategoryName:              row.CategoryName,
			CategoryGroupID:           row.CategoryGroupID,
			CategoryGroupName:         row.CategoryGroupName,
			DestAccountID:             row.TrxLine.DestAccountID,
			DestAccountName:           row.DestAccountName,
			ExpenseShareID:            row.TrxLine.ExpenseShareID,
			ExpenseShareDefaultName:   row.ExpenseShareDefaultName,
			SplitBudgetExpenseShareID: row.TrxLine.SplitBudgetExpenseShareID,
			SplitBudgetExpenseShareName: row.SplitBudgetExpenseShareName,
			SplitBudgetExpenseShareDisplayName: row.SplitBudgetExpenseShareDisplayName,
			SplitBudgetID:             row.SplitBudgetID,
			SplitBudgetName:           row.SplitBudgetName,
			SplitBudgetLoginID:        row.SplitBudgetLoginID,
			DestBudgetExpenseShareID:  row.TrxLine.DestBudgetExpenseShareID,
			DestBudgetExpenseShareName: row.DestBudgetExpenseShareName,
			DestBudgetExpenseShareDisplayName: row.DestBudgetExpenseShareDisplayName,
			DestBudgetID:              row.DestBudgetID,
			DestBudgetName:            row.DestBudgetName,
			DestBudgetLoginID:         row.DestBudgetLoginID,
			Reconciled:                t.Reconciled(),
		}, t)
		return line, nil
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

// The transaction line expense share, returns an error if its not a split
// or settlement line.
func (l *TrxLine) Share() (*ExpenseShare, error) {
	if l.row.ExpenseShareID.Valid {
		return l.share, nil
	}
	return nil, fmt.Errorf("Transaction line is not a split or settlement line")
}

// The split line tagged member budget, returns an error if its not a split.
func (l *TrxLine) SplitBudget() (*ExpenseShareMembership, error) {
	if l.splitBudget != nil {
		return l.splitBudget, nil
	}
	return nil, fmt.Errorf("Transaction line is not a split line")
}

// The settlement line counterparty member budget, returns an error if its
// not a settlement.
func (l *TrxLine) DestBudget() (*ExpenseShareMembership, error) {
	if l.destBudget != nil {
		return l.destBudget, nil
	}
	return nil, fmt.Errorf("Transaction line is not a settlement line")
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

func (l *TrxLine) Update(ctx context.Context, destAccount *Account, category *Category, income bool, share *ExpenseShare, splitBudget, destBudget *ExpenseShareMembership, outflow, inflow int) error {
	if destAccount != nil && destAccount.ID() == l.trx.account.ID() {
		return fmt.Errorf("Transfer destination must differ from source account")
	}
	defer cache.InvalidateResults(ctx)
	destAccountID := sql.NullInt64{}
	if destAccount != nil {
		destAccountID = sql.NullInt64{Valid: true, Int64: int64(destAccount.ID())}
	}
	categoryID := sql.NullInt64{}
	if category != nil {
		categoryID = sql.NullInt64{Valid: true, Int64: int64(category.ID())}
	}
	shareID := sql.NullInt64{}
	if share != nil {
		shareID = sql.NullInt64{Valid: true, Int64: int64(share.ID())}
	}
	splitBudgetID := sql.NullInt64{}
	if splitBudget != nil {
		splitBudgetID = sql.NullInt64{Valid: true, Int64: int64(splitBudget.ID())}
	}
	destBudgetID := sql.NullInt64{}
	if destBudget != nil {
		destBudgetID = sql.NullInt64{Valid: true, Int64: int64(destBudget.ID())}
	}
	row, err := l.trx.account.budget.service.repo.UpdateTrxLine(ctx, data.UpdateTrxLineParams{
		ID:             int64(l.ID()),
		TrxID:          int64(l.trx.ID()),
		BudgetID:       int64(l.trx.account.budget.ID()),
		LoginID:        int64(l.trx.account.budget.LoginID()),
		DestAccountID:  destAccountID,
		CategoryID:     categoryID,
		Income:         income,
		ExpenseShareID: shareID,
		SplitBudgetExpenseShareID:  splitBudgetID,
		DestBudgetExpenseShareID:   destBudgetID,
		Outflow:        int64(outflow),
		Inflow:         int64(inflow),
	})
	if err != nil {
		return err
	}
	l.row = row
	l.share, l.splitBudget, l.destBudget = share, splitBudget, destBudget
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
