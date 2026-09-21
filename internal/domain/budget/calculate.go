package budget

import (
	"context"
	"time"
)

type MonthSummary struct {
	// Total allocated in the month
	Allocated int
	// The total amount of money available for the month
	// For current of past months, it's the sum of all the allocations minus the spending
	// For future months it's the sum of allocations
	Available int
	// The money that is available to assign to categories. This is net worth minus all allocations
	ReadyToAssign int
	// Sum of all income transaction lines
	Income int
	// Sum of all categorized transaction lines
	Spent int
	// The difference of all transactions totals and the line totals
	Uncategorized int
}

func (b *Budget) GetMonthSummary(ctx context.Context, month time.Time) (MonthSummary, error) {
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-1 * time.Nanosecond)
	summary := MonthSummary{}
	trxs, err := b.trxService.List(ctx, int(b.row.LoginID), int(b.row.ID))
	if err != nil {
		return summary, err
	}
	allocations, err := b.allocationService.List(ctx, int(b.row.LoginID), int(b.row.ID))
	if err != nil {
		return summary, err
	}
	for _, trx := range trxs {
		categorized := 0
		for _, line := range trx.Lines() {
			categorized += line.Inflow() - line.Outflow()
			if timeInsideMonth(trx.Date(), startOfMonth, endOfMonth) {
				if line.IsIncome() {
					summary.Income += line.Inflow()
				}
				if _, err := line.Category(); err == nil {
					summary.Spent += line.Inflow() - line.Outflow()
				}
			}
		}
		summary.ReadyToAssign += trx.TotalInflow() - trx.TotalOutflow()
		summary.Uncategorized += trx.TotalInflow() - trx.TotalOutflow() - categorized
	}
	futureMonth := time.Now().Before(startOfMonth)
	for _, a := range allocations {
		if timeInsideMonth(a.Month(), startOfMonth, endOfMonth) {
			summary.Available += a.Amount()
			summary.Allocated += a.Amount()
		} else if !futureMonth && a.Month().Before(endOfMonth) {
			summary.Available += a.Amount()
		}
		summary.ReadyToAssign -= a.Amount()
	}
	summary.Available += summary.Spent
	return summary, nil
}

func timeInsideMonth(t, start, end time.Time) bool {
	return t.Equal(start) || t.Equal(end) || (t.After(start) && t.Before(end))
}
