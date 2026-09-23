package domain

import (
	"context"
	"fmt"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type TrxLine struct {
	service            *TrxService
	row                data.TrxLine
	category           *Category
	destinationAccount *Account
}

func (l *TrxLine) ID() int {
	return int(l.row.ID)
}

func (l *TrxLine) IsIncome() bool {
	return l.row.Income
}

func (l *TrxLine) Category() (*Category, error) {
	if l.row.CategoryID.Valid {
		return l.category, nil
	}
	return nil, fmt.Errorf("Transaction line is not a categorization")
}

func (l *TrxLine) DestinationAccount() (*Account, error) {
	if l.row.DestAccountID.Valid {
		return l.destinationAccount, nil
	}
	return nil, fmt.Errorf("Transaction line is not a transfer")
}

func (l *TrxLine) Inflow() int {
	return int(l.row.Inflow)
}

func (l *TrxLine) Outflow() int {
	return int(l.row.Outflow)
}

func (l *TrxLine) TrxID() int {
	return int(l.row.TrxID)
}

func (l *TrxLine) Update(ctx context.Context, loginID, trxID int, destAccountID, categoryID *int, income bool, outflow, inflow int) error {
	defer cache.InvalidateResults(ctx)
	row, err := l.service.repo.UpdateTrxLine(ctx, data.UpdateTrxLineParams{
		TrxID:         int64(trxID),
		DestAccountID: nullInt64FromInt(destAccountID),
		CategoryID:    nullInt64FromInt(categoryID),
		Income:        income,
		Outflow:       int64(outflow),
		Inflow:        int64(inflow),
		ID:            l.row.ID,
		BudgetID:      l.row.BudgetID,
		LoginID:       int64(loginID),
	})
	if err != nil {
		return err
	}
	l.row = row
	// Keep the refs in sync with the row (nil when the id is null/cleared).
	l.destinationAccount = nil
	if destAccountID != nil {
		dest, err := l.service.accountService.Get(ctx, loginID, int(l.row.BudgetID), *destAccountID)
		if err != nil {
			return err
		}
		l.destinationAccount = dest
	}
	l.category = nil
	if categoryID != nil {
		category, err := l.service.categoryService.Get(ctx, loginID, int(l.row.BudgetID), *categoryID)
		if err != nil {
			return err
		}
		l.category = category
	}
	return nil
}

func (l *TrxLine) Delete(ctx context.Context, loginID int) error {
	defer cache.InvalidateResults(ctx)
	return l.service.repo.DeleteTrxLine(ctx, data.DeleteTrxLineParams{
		ID:       l.row.ID,
		BudgetID: l.row.BudgetID,
		LoginID:  int64(loginID),
	})
}

