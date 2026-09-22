package domain

import (
	"context"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type Goal struct {
	service           *GoalService
	allocationService *AllocationService
	row               data.Goal
}

func (s *GoalService) fromRow(ctx context.Context, row data.Goal) *Goal {
	if v, ok := cache.Get[*Goal](ctx, row.ID); ok {
		return v
	}
	goal := &Goal{
		service:           s,
		allocationService: s.allocationService,
		row:               row,
	}
	cache.Store(ctx, row.ID, goal)
	return goal
}

func (g *Goal) Category() int {
	return int(g.row.CategoryID)
}

func (g *Goal) Type() string {
	return g.row.Type
}

func (g *Goal) StartDate() time.Time {
	return g.row.StartDate.Time
}

func (g *Goal) EndDate() *time.Time {
	if g.row.EndDate.Valid {
		return &g.row.EndDate.Time
	} else {
		return nil
	}
}

func (g *Goal) Amount() int {
	return int(g.row.Amount)
}

type GoalValues struct {
	AllocatedToDate int
	NeededForMonth  int
	Gap int
}

func (g *Goal) GoalValues(ctx context.Context, loginID int, month time.Time) (*GoalValues, error) {
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-1 * time.Nanosecond)
	if g.StartDate().After(startOfMonth) || (g.EndDate() != nil && g.EndDate().Before(startOfMonth)) {
		return nil, nil
	}
	allocations, err := g.allocationService.List(ctx, loginID, int(g.row.BudgetID))
	if err != nil {
		return nil, err
	}
	values := GoalValues{}
	allocatedThisMonth := 0
	for _, allocation := range allocations {
		if allocation.Category() != g.Category() {
			continue
		}
		if allocation.Month().Equal(startOfMonth) {
			allocatedThisMonth = allocation.Amount()
		}
		if allocation.Month().Before(endOfMonth) {
			values.AllocatedToDate += allocation.Amount()
		}
	}
	switch g.Type() {
	case "monthly":
		values.NeededForMonth = g.Amount()
	case "save":
		years := g.EndDate().Year() - startOfMonth.Year()
		months := int(g.EndDate().Month()) - int(startOfMonth.Month())
		monthsLeft :=  years*12 + months + 1
		values.NeededForMonth = (g.Amount() - values.AllocatedToDate + allocatedThisMonth) / monthsLeft
	case "refill":
		values.NeededForMonth = g.Amount()
	}
	values.Gap = values.NeededForMonth - allocatedThisMonth
	return &values, nil
}
