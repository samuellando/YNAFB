package domain

import (
	"context"
	"fmt"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/db/types"
)

type Allocation struct {
	row      data.Allocation
	category *Category
}

// Create an Allocation object from a data row.
func allocationFromRow(ctx context.Context, row data.Allocation, category *Category) *Allocation {
	allocation, _ := cache.Get(ctx, row.ID, func() (*Allocation, error) {
		return &Allocation{
			row:      row,
			category: category,
		}, nil
	})
	return allocation
}

// List all the allocations in the budget
func (b *Budget) ListAllocations(ctx context.Context) ([]*Allocation, error) {
	return cache.Result(ctx, fmt.Sprintf("allocationServiceList-%d-%d", b.LoginID(), b.ID()), func() ([]*Allocation, error) {
		rows, err := b.service.repo.ListAllocations(ctx, data.ListAllocationsParams{
			BudgetID: int64(b.ID()),
			LoginID:  int64(b.LoginID()),
		})
		if err != nil {
			return nil, err
		}
		allocations := make([]*Allocation, len(rows))
		for i, row := range rows {
			var group *CategoryGroup
			if row.CategoryGroupID.Valid {
				group = categoryGroupFromRow(ctx, data.CategoryGroup{
					ID:       row.CategoryGroupID.Int64,
					BudgetID: row.BudgetID,
					Name:     row.CategoryGroupName.String,
				}, b)
			}
			category := categoryFromRow(ctx, data.Category{
				ID:              row.CategoryID,
				BudgetID:        row.BudgetID,
				Name:            row.CategoryName,
				CategoryGroupID: row.CategoryGroupID,
			}, b, group)
			allocations[i] = allocationFromRow(ctx, data.Allocation{
				ID:         row.ID,
				BudgetID:   row.BudgetID,
				CategoryID: row.CategoryID,
				Month:      row.Month,
				Amount:     row.Amount,
			}, category)
		}
		return allocations, nil
	})
}

// Set the allocation for a category in a month
func (c *Category) SetAllocation(ctx context.Context, month time.Time, amount int) (*Allocation, error) {
	defer cache.InvalidateResults(ctx)
	row, err := c.budget.service.repo.SetAllocation(ctx, data.SetAllocationParams{
		CategoryID: int64(c.ID()),
		Month:      types.UnixTime{Time: month},
		Amount:     int64(amount),
		BudgetID:   int64(c.budget.ID()),
		LoginID:    int64(c.budget.LoginID()),
	})
	if err != nil {
		return nil, err
	}
	allocation := allocationFromRow(ctx, row, c)
	return allocation, nil
}

// Get the allocation's ID
func (a *Allocation) ID() int {
	return int(a.row.ID)
}

// Get the allocation's category
func (a *Allocation) Category() *Category {
	return a.category
}

// Get the allocation's month
func (a *Allocation) Month() time.Time {
	return a.row.Month.Time
}

// Get the allocation's amount
func (a *Allocation) Amount() int {
	return int(a.row.Amount)
}
