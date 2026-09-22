package domain

import (
	"context"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type PayeeDefaultLine struct {
	service *PayeeService
	row     data.PayeeDefaultLine
}

func (s *PayeeService) defaultLineFromRow(ctx context.Context, row data.PayeeDefaultLine) *PayeeDefaultLine {
	if cached, ok := cache.Get[*PayeeDefaultLine](ctx, row.ID); ok {
		return cached
	}
	line := &PayeeDefaultLine{
		service: s,
		row:     row,
	}
	cache.Store(ctx, row.ID, line)
	return line
}

func (l *PayeeDefaultLine) ID() int {
	return int(l.row.ID)
}

func (l *PayeeDefaultLine) PayeeID() int {
	return int(l.row.PayeeID)
}

func (l *PayeeDefaultLine) DestAccountID() *int {
	if !l.row.DestAccountID.Valid {
		return nil
	}
	id := int(l.row.DestAccountID.Int64)
	return &id
}

func (l *PayeeDefaultLine) CategoryID() *int {
	if !l.row.CategoryID.Valid {
		return nil
	}
	id := int(l.row.CategoryID.Int64)
	return &id
}

func (l *PayeeDefaultLine) Income() bool {
	return l.row.Income
}

func (l *PayeeDefaultLine) Percent() int {
	return int(l.row.Percent)
}

// Update rewrites the line. loginID is passed explicitly because,
// unlike Account, PayeeDefaultLine holds no budget reference to derive it from.
func (l *PayeeDefaultLine) Update(ctx context.Context, loginID, payeeID int, destAccountID, categoryID *int, income bool, percent int) error {
	defer cache.InvalidateResults(ctx)
	row, err := l.service.repo.UpdatePayeeDefaultLine(ctx, data.UpdatePayeeDefaultLineParams{
		PayeeID:       int64(payeeID),
		DestAccountID: nullInt64FromInt(destAccountID),
		CategoryID:    nullInt64FromInt(categoryID),
		Income:        income,
		Percent:       int64(percent),
		ID:            l.row.ID,
		BudgetID:      l.row.BudgetID,
		LoginID:       int64(loginID),
	})
	if err != nil {
		return err
	}
	l.row = row
	return nil
}

// Delete removes the line. loginID is passed explicitly, see Update.
func (l *PayeeDefaultLine) Delete(ctx context.Context, loginID int) error {
	defer cache.InvalidateResults(ctx)
	return l.service.repo.DeletePayeeDefaultLine(ctx, data.DeletePayeeDefaultLineParams{
		ID:       l.row.ID,
		BudgetID: l.row.BudgetID,
		LoginID:  int64(loginID),
	})
}
