package budget

import (
	"context"
	"slices"
	"strings"
	"time"

	"samuellando.com/YNAFB/internal/domain/category"
	"samuellando.com/YNAFB/internal/domain/goal"
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

type MonthCategory struct {
	ID        int
	Name      string
	Group     *category.Group
	Allocated int
	Spent     int
	CarryOver int
	Available int
	Goal      *goal.Goal
}

func (b *Budget) GetMonthSummary(ctx context.Context, month time.Time) (MonthSummary, error) {
	startOfMonth, endOfMonth := getStartAndEndOfMonth(month)
	summary := MonthSummary{}
	fetchedData, err := b.getCalculationData(ctx)
	if err != nil {
		return summary, err
	}
	trxs := fetchedData.trxs
	allocations := fetchedData.allocations
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

func (b *Budget) GetMonthCategories(ctx context.Context, month time.Time) ([]*MonthCategory, error) {
	startOfMonth, endOfMonth := getStartAndEndOfMonth(month)
	futureMonth := time.Now().Before(startOfMonth)
	data, err := b.getCalculationData(ctx)
	if err != nil {
		return nil, err
	}
	categoriesMap := make(map[int]*MonthCategory)
	for _, category := range data.categories {
		categoriesMap[category.ID()] = &MonthCategory{
			ID:    category.ID(),
			Name:  category.Name(),
			Group: category.Group(),
		}
	}
	for _, trx := range data.trxs {
		for _, line := range trx.Lines() {
			if category, err := line.Category(); err == nil {
				if timeInsideMonth(trx.Date(), startOfMonth, endOfMonth) {
					categoriesMap[category.ID()].Spent += line.Inflow() - line.Outflow()
					categoriesMap[category.ID()].Available += line.Inflow() - line.Outflow()
				} else if !futureMonth && trx.Date().Before(endOfMonth) {
					categoriesMap[category.ID()].Available += line.Inflow() - line.Outflow()
					categoriesMap[category.ID()].CarryOver += line.Inflow() - line.Outflow()
				}
			}
		}
	}
	for _, allocation := range data.allocations {
		if timeInsideMonth(allocation.Month(), startOfMonth, endOfMonth) {
			categoriesMap[allocation.Category()].Allocated = allocation.Amount()
			categoriesMap[allocation.Category()].Available += allocation.Amount()
		} else if !futureMonth && allocation.Month().Before(endOfMonth) {
			categoriesMap[allocation.Category()].Available += allocation.Amount()
			categoriesMap[allocation.Category()].CarryOver += allocation.Amount()
		}
	}
	for _, goal := range data.goals {
		if goal.StartDate().Before(startOfMonth) || goal.StartDate().Equal(startOfMonth) {
			if goal.EndDate() == nil || goal.EndDate().After(startOfMonth) || goal.EndDate().Equal(startOfMonth) {
				categoriesMap[goal.Category()].Goal = goal
			}
		}
	}
	categories := make([]*MonthCategory, 0)
	for _, cat := range categoriesMap {
		categories = append(categories, cat)
	}
	slices.SortFunc(categories, func(a, b *MonthCategory) int {
		o := strings.Compare(a.Name, b.Name)
		if a.Group == nil && b.Group == nil {
			return o
		}
		if a.Group == nil {
			return -1
		} else if b.Group == nil {
			return 1
		} else {
			if og := strings.Compare(a.Group.Name(), b.Group.Name()); og != 0 {
				return og
			} else {
				return o
			}
		}
	})
	return categories, nil
}
