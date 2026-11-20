package domain

import (
	"context"
	"slices"
	"strings"
	"time"
)

var zero time.Time = time.Unix(0, 0)

type MonthSummary struct {
	// Total allocated in the month
	Allocated int
	// The money that is available to assign to categories.
	// Income - spend to date minus any overspending from previous months
	ReadyToAssign int
	// Sum of all income transaction lines this month
	Income int
	// Sum of all categorized transaction lines this month
	Spent int
	// The difference of all transactions totals and the line totals
	Uncategorized int
}

type MonthCategory struct {
	ID        int
	Name      string
	Group     *CategoryGroup
	Allocated int
	Spent     int
	Available int
	Goal      *Goal
}

type fetchedData struct {
	trxs                            []*Trx
	allocations                     []*Allocation
	goals                           []*Goal
	categories                      []*Category
	shares                          []*ExpenseShareMembership
	shareTransactions               map[int][]*ExpenseShareTransaction
	shareTransactionSplitLines      map[int][]*TrxLine
	shareTransactionCategorizations map[int][]*SplitCategorization
}

func (b *Budget) getCalculationData(ctx context.Context) (*fetchedData, error) {
	trxs, err := b.listTransactions(ctx)
	if err != nil {
		return nil, err
	}
	allocations, err := b.ListAllocations(ctx)
	if err != nil {
		return nil, err
	}
	goals, err := b.ListGoals(ctx)
	if err != nil {
		return nil, err
	}
	categories, err := b.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	shares, err := b.ListExpenseShares(ctx)
	if err != nil {
		return nil, err
	}
	shareTransactions := make(map[int][]*ExpenseShareTransaction)
	shareTransactionsCategorizations := make(map[int][]*SplitCategorization)
	shareTransactionSplitLines := make(map[int][]*TrxLine)
	for _, share := range shares {
		trxs, err := share.ListTransactions(ctx)
		if err != nil {
			return nil, err
		}
		for _, trx := range trxs {
			cats, err := trx.MyCategorizations(ctx)
			if err != nil {
				return nil, err
			}
			shareTransactionsCategorizations[trx.SourceTrx().ID()] = cats
			lines, err := trx.MySplitLines(ctx)
			if err != nil {
				return nil, err
			}
			shareTransactionSplitLines[trx.SourceTrx().ID()] = lines
		}
		shareTransactions[share.share.ID()] = trxs
	}
	return &fetchedData{
		trxs:                            trxs,
		allocations:                     allocations,
		goals:                           goals,
		categories:                      categories,
		shares:                          shares,
		shareTransactions:               shareTransactions,
		shareTransactionSplitLines:      shareTransactionSplitLines,
		shareTransactionCategorizations: shareTransactionsCategorizations,
	}, nil
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
	// Accumulator variables
	incomeBeforeMonth := 0
	incomeThisMonth := 0
	spentBeforeMonth := 0
	spentThisMonth := 0
	settledBeforeMonth := 0
	settledThisMonth := 0
	uncategorized := 0
	allocatedThisMonth := 0
	expenseShareLoaned := 0
	expenseShareDebt := 0
	// Itterate over all transactions and allocations and accumulate
	for _, trx := range trxs {
		categorized := 0
		for _, line := range trx.Lines() {
			lineAmount := line.Inflow() - line.Outflow()
			categorized += lineAmount
			if trx.Date().Before(startOfMonth) {
				if _, err := line.Category(); err == nil {
					spentBeforeMonth -= lineAmount
				}
				if line.IsIncome() {
					incomeBeforeMonth += lineAmount
				}
				if line.IsSplit() {
					expenseShareLoaned -= lineAmount
				}
				if line.IsSettlement() {
					settledBeforeMonth -= lineAmount
				}
			} else if timeInsideMonth(trx.Date(), startOfMonth, endOfMonth) {
				if _, err := line.Category(); err == nil {
					spentThisMonth -= lineAmount
				}
				if line.IsIncome() {
					incomeThisMonth += lineAmount
				}
				if line.IsSplit() {
					expenseShareLoaned -= lineAmount
				}
				if line.IsSettlement() {
					settledThisMonth -= lineAmount
				}
			}
		}
		uncategorized += abs(trx.TotalInflow() - trx.TotalOutflow() - categorized)
	}
	// Splits in expense shares count as spending as well
	for _, cs := range fetchedData.shareTransactionCategorizations {
		categorized := 0
		totalIn := 0
		totalOut := 0
		for trxID, c := range cs {
			for _, s := range fetchedData.shareTransactionSplitLines[trxID] {
				totalIn += s.Inflow()
				totalOut += s.Outflow()
			}
			lineAmount := c.Inflow() - c.Outflow()
			expenseShareDebt -= lineAmount
			if c.Transaction().Date().Before(startOfMonth) {
				spentBeforeMonth -= lineAmount
			} else if timeInsideMonth(c.Transaction().Date(), startOfMonth, endOfMonth) {
				spentThisMonth -= lineAmount
			}
		}
		uncategorized += abs(totalIn - totalOut - categorized)
	}
	// Calculate allocation for this month
	for _, a := range allocations {
		if timeInsideMonth(a.Month(), startOfMonth, endOfMonth) {
			allocatedThisMonth += a.Amount()
		}
	}
	// Intermediate calculations
	incomeToEndOfMonth := incomeBeforeMonth + incomeThisMonth
	settlementsToEndOfMonth := settledBeforeMonth + settledThisMonth
	overSettled := 0
	if settlementsToEndOfMonth > expenseShareDebt {
		overSettled = settlementsToEndOfMonth - expenseShareDebt
	}
	// Set the final values
	summary.Allocated = allocatedThisMonth
	summary.Income = incomeThisMonth
	summary.Spent = spentThisMonth
	summary.Uncategorized = uncategorized
	summary.ReadyToAssign = (incomeToEndOfMonth - spentBeforeMonth - allocatedThisMonth - expenseShareLoaned - overSettled) - calculateCarryIn(fetchedData, month)
	return summary, nil
}

func (b *Budget) GetMonthCategories(ctx context.Context, month time.Time) ([]*MonthCategory, error) {
	startOfMonth := getStartOfMonth(month)
	futureMonth := time.Now().Before(startOfMonth)
	data, err := b.getCalculationData(ctx)
	if err != nil {
		return nil, err
	}
	categoryValues := calculateMonthCategoryValues(data, startOfMonth)
	categoriesMap := make(map[int]*MonthCategory)
	for _, category := range data.categories {
		group, _ := category.Group()
		v := &MonthCategory{
			ID:        category.ID(),
			Name:      category.Name(),
			Group:     group,
			Spent:     categoryValues[category.ID()].spent,
			Allocated: categoryValues[category.ID()].allocated,
			Available: categoryValues[category.ID()].available,
		}
		// Suppress any carry over for future months so allocations can be made conservatively
		if futureMonth {
			v.Available = v.Allocated - v.Spent
		}
		categoriesMap[category.ID()] = v
	}
	// Determine the active goals for the month
	for _, goal := range data.goals {
		if goal.StartDate().Before(startOfMonth) || goal.StartDate().Equal(startOfMonth) {
			if goal.EndDate() == nil || goal.EndDate().After(startOfMonth) || goal.EndDate().Equal(startOfMonth) {
				categoriesMap[goal.Category().ID()].Goal = goal
			}
		}
	}
	// COnvert to a slice and sort
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

type monthTotals struct {
	allocated int
	spent     int
	available int
}

func calculateMonthCategoryValues(data *fetchedData, month time.Time) map[int]*monthTotals {
	// Accumulate the monthly spending and allocations
	minMonth, _, monthlyTotals := aggregateCategoryMonthlyEntrties(data)
	// Convert to an contiguous timeline array
	res := make(map[int]*monthTotals)
	// Init the result
	for _, category := range data.categories {
		res[category.ID()] = &monthTotals{
			spent:     0,
			allocated: 0,
			available: 0,
		}
	}
	// No data, return zeros
	if minMonth.Equal(zero) {
		return res
	}
	for iMonth := minMonth; timeLE(iMonth, month); iMonth = iMonth.AddDate(0, 1, 0) {
		for _, category := range data.categories {
			v, ok := monthlyTotals[category.ID()][iMonth]
			if !ok {
				v = &monthTotals{
					spent:     0,
					allocated: 0,
					available: 0,
				}
			}
			res[category.ID()].spent = v.spent
			res[category.ID()].allocated = v.allocated
			// Only carry forward possitive available amounts. Negatives
			// will be subtracted from the ready to assign for next month
			if pa := res[category.ID()].available; pa > 0 {
				res[category.ID()].available = v.allocated - v.spent + pa
			} else {
				res[category.ID()].available = v.allocated - v.spent
			}
		}
	}
	return res
}

func calculateCarryIn(data *fetchedData, month time.Time) int {
	monthlyValues := calculateMonthCategoryValues(data, month.AddDate(0, -1, 0))
	available := 0
	for _, catValues := range monthlyValues {
		if catValues.available > 0 {
			available += catValues.available
		}
	}
	return available
}

func aggregateCategoryMonthlyEntrties(data *fetchedData) (time.Time, time.Time, map[int]map[time.Time]*monthTotals) {
	var minMonth time.Time = zero
	var maxMonth time.Time = zero
	// Accumulate the monthly spending and allocations
	monthlyTotals := make(map[int]map[time.Time]*monthTotals)
	for _, category := range data.categories {
		monthlyTotals[category.ID()] = make(map[time.Time]*monthTotals)
	}
	for _, trx := range data.trxs {
		for _, line := range trx.Lines() {
			lineAmount := line.Inflow() - line.Outflow()
			if cat, err := line.Category(); err == nil {
				month := getStartOfMonth(trx.Date())
				minMonth, maxMonth = timeMinMax(minMonth, maxMonth, month)
				if m, ok := monthlyTotals[cat.ID()][month]; ok {
					m.spent -= lineAmount
				} else {
					monthlyTotals[cat.ID()][month] = &monthTotals{
						spent: -1 * lineAmount,
					}
				}
			}
		}
	}
	for _, cats := range data.shareTransactionCategorizations {
		for _, cat := range cats {
			lineAmount := cat.Inflow() - cat.Outflow()
			month := getStartOfMonth(cat.Transaction().Date())
			minMonth, maxMonth = timeMinMax(minMonth, maxMonth, month)
			if m, ok := monthlyTotals[cat.Category().ID()][month]; ok {
				m.spent -= lineAmount
			} else {
				monthlyTotals[cat.Category().ID()][month] = &monthTotals{
					spent: -1 * lineAmount,
				}
			}
		}
	}
	for _, a := range data.allocations {
		month := getStartOfMonth(a.Month())
		minMonth, maxMonth = timeMinMax(minMonth, maxMonth, month)
		if m, ok := monthlyTotals[a.category.ID()][month]; ok {
			m.allocated += a.Amount()
		} else {
			monthlyTotals[a.category.ID()][month] = &monthTotals{
				allocated: a.Amount(),
			}
		}
	}
	return minMonth, maxMonth, monthlyTotals
}

func timeMinMax(a, b, c time.Time) (time.Time, time.Time) {
	zero := time.Unix(0, 0)
	var minT time.Time
	var maxT time.Time
	if !a.Equal(zero) {
		minT = a
		maxT = a
	} else if !b.Equal(zero) {
		minT = b
		maxT = b
	} else {
		return c, c
	}
	if b.Before(minT) {
		minT = b
	}
	if c.Before(minT) {
		minT = c
	}
	if b.After(maxT) {
		maxT = b
	}
	if c.After(maxT) {
		maxT = c
	}
	return minT, maxT
}
