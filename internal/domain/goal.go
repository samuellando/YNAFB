package domain

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"samuellando.com/YNAFB/internal/data"
	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/db/types"
)

type Goal struct {
	row      data.Goal
	category *Category
}

// Load a goal from a data row
func goalFromRow(ctx context.Context, row data.ListGoalsRow, budget *Budget) *Goal {
	goal, _ := cache.Get(ctx, row.ID, func() (*Goal, error) {
		var categoryGroup *CategoryGroup
		if row.CategoryGroupID.Valid {
			categoryGroup = categoryGroupFromRow(ctx, data.CategoryGroup{
				ID:       row.CategoryGroupID.Int64,
				BudgetID: int64(budget.ID()),
				Name:     row.CategoryGroupName.String,
			}, budget)
		}
		category := categoryFromRow(ctx, data.Category{
			ID:              row.CategoryID,
			Name:            row.CategoryName,
			BudgetID:        row.BudgetID,
			CategoryGroupID: row.CategoryGroupID,
		}, budget, categoryGroup)
		return &Goal{
			row: data.Goal{
				ID:         row.ID,
				BudgetID:   int64(budget.ID()),
				CategoryID: row.CategoryID,
				Type:       row.Type,
				StartDate:  row.StartDate,
				EndDate:    row.EndDate,
				Amount:     row.Amount,
			},
			category: category,
		}, nil
	})
	return goal
}

// List all the goals in the budget
func (b *Budget) ListGoals(ctx context.Context) ([]*Goal, error) {
	return cache.Result(ctx, fmt.Sprintf("goalServiceList-%d-%d", b.LoginID(), b.ID()), func() ([]*Goal, error) {
		rows, err := b.service.repo.ListGoals(ctx, data.ListGoalsParams{
			BudgetID: int64(b.ID()),
			LoginID:  int64(b.LoginID()),
		})
		if err != nil {
			return nil, err
		}
		goals := make([]*Goal, len(rows))
		for i, row := range rows {
			goals[i] = goalFromRow(ctx, row, b)
		}
		return goals, nil
	})
}

// Get the goal fro a category
func (c *Category) GetGoal(ctx context.Context) (*Goal, error) {
	row, err := c.budget.service.repo.GetGoalByCategory(ctx, data.GetGoalByCategoryParams{
		LoginID:    int64(c.budget.LoginID()),
		BudgetID:   int64(c.budget.ID()),
		CategoryID: int64(c.ID()),
	})
	if err != nil {
		return nil, err
	}
	categoryGroupID := sql.NullInt64{}
	categoryGroupName := sql.NullString{}
	if group, err := c.Group(); err == nil {
		categoryGroupID = sql.NullInt64{Valid: true, Int64: int64(group.ID())}
		categoryGroupName = sql.NullString{Valid: true, String: group.Name()}
	}
	goal := goalFromRow(ctx, data.ListGoalsRow{
		ID:                   row.ID,
		BudgetID:             row.BudgetID,
		CategoryID:           row.CategoryID,
		Type:                 row.Type,
		StartDate:            row.StartDate,
		EndDate:              row.EndDate,
		Amount:               row.Amount,
		CategoryName:         c.Name(),
		CategoryGroupID:      categoryGroupID,
		CategoryGroupName: categoryGroupName,
	}, c.budget)
	return goal, nil
}

// Create a goal in a category
func (c *Category) Create(ctx context.Context, goalType string, start time.Time, end *time.Time, amount int) (*Goal, error) {
	defer cache.InvalidateResults(ctx)
	endTime := types.NullUnixTime{}
	if end != nil {
		endTime = types.NullUnixTime{Valid: true, Time: *end}
	}
	row, err := c.budget.service.repo.CreateGoal(ctx, data.CreateGoalParams{
		Type:       goalType,
		StartDate:  types.UnixTime{Time: start},
		EndDate:    endTime,
		CategoryID: int64(c.ID()),
		Amount:     int64(amount),
		BudgetID:   int64(c.budget.ID()),
		LoginID:    int64(c.budget.LoginID()),
	})
	if err != nil {
		return nil, err
	}
	categoryGroupID := sql.NullInt64{}
	categoryGroupName := sql.NullString{}
	if group, err := c.Group(); err == nil {
		categoryGroupID = sql.NullInt64{Valid: true, Int64: int64(group.ID())}
		categoryGroupName = sql.NullString{Valid: true, String: group.Name()}
	}
	goal := goalFromRow(ctx, data.ListGoalsRow{
		ID:                   row.ID,
		BudgetID:             row.BudgetID,
		CategoryID:           row.CategoryID,
		Type:                 row.Type,
		StartDate:            row.StartDate,
		EndDate:              row.EndDate,
		Amount:               row.Amount,
		CategoryName:         c.Name(),
		CategoryGroupID:      categoryGroupID,
		CategoryGroupName: categoryGroupName,
	}, c.budget)
	return goal, nil
}

// Get the goal ID
func (g *Goal) ID() int {
	return int(g.row.ID)
}

// Get the goal category
func (g *Goal) Category() *Category {
	return g.category
}

// Get the goal type
func (g *Goal) Type() string {
	return g.row.Type
}

// Get the goal start date
func (g *Goal) StartDate() time.Time {
	return g.row.StartDate.Time
}

// Get the goal end date
func (g *Goal) EndDate() *time.Time {
	if g.row.EndDate.Valid {
		return &g.row.EndDate.Time
	} else {
		return nil
	}
}

// Get the goal amount
func (g *Goal) Amount() int {
	return int(g.row.Amount)
}

// Update the goal
func (g *Goal) Update(ctx context.Context, goalType string, start time.Time, end *time.Time, amount int) error {
	defer cache.InvalidateResults(ctx)
	endTime := types.NullUnixTime{}
	if end != nil {
		endTime = types.NullUnixTime{Valid: true, Time: *end}
	}
	row, err := g.category.budget.service.repo.UpdateGoal(ctx, data.UpdateGoalParams{
		Type:       goalType,
		StartDate:  types.UnixTime{Time: start},
		EndDate:    endTime,
		Amount:     int64(amount),
		BudgetID:   int64(g.category.budget.ID()),
		LoginID:    int64(g.category.budget.LoginID()),
		CategoryID: g.row.CategoryID,
	})
	if err != nil {
		return err
	}
	g.row = row
	return nil
}

// Delete a goal
func (g *Goal) Delete(ctx context.Context) error {
	defer cache.InvalidateResults(ctx)
	defer cache.Delete[*Goal](ctx, int64(g.ID()))
	return g.category.budget.service.repo.DeleteGoal(ctx, data.DeleteGoalParams{
		BudgetID:   int64(g.category.budget.ID()),
		LoginID:    int64(g.category.budget.LoginID()),
		CategoryID: g.row.CategoryID,
	})
}

type GoalValues struct {
	AllocatedToDate int
	NeededForMonth  int
	Gap             int
}

// Compute values for a goal
func (g *Goal) GoalValues(ctx context.Context, month time.Time) (*GoalValues, error) {
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-1 * time.Nanosecond)
	if g.StartDate().After(startOfMonth) || (g.EndDate() != nil && g.EndDate().Before(startOfMonth)) {
		return nil, nil
	}
	allocations, err := g.category.budget.ListAllocations(ctx)
	if err != nil {
		return nil, err
	}
	values := GoalValues{}
	allocatedThisMonth := 0
	for _, allocation := range allocations {
		if allocation.category.ID() != g.category.ID() {
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
		monthsLeft := years*12 + months + 1
		values.NeededForMonth = (g.Amount() - values.AllocatedToDate + allocatedThisMonth) / monthsLeft
	case "refill":
		values.NeededForMonth = g.Amount()
	}
	values.Gap = values.NeededForMonth - allocatedThisMonth
	return &values, nil
}
