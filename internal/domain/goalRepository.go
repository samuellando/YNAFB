package domain

import (
	"context"
	"samuellando.com/YNAFB/data"
)

type GoalRepository interface {
	ListGoals(context.Context, data.ListGoalsParams) ([]data.Goal, error)
	GetGoalByCategory(context.Context, data.GetGoalByCategoryParams) (data.GetGoalByCategoryRow, error)
	CreateGoal(context.Context, data.CreateGoalParams) (data.Goal, error)
	UpdateGoal(context.Context, data.UpdateGoalParams) (data.Goal, error)
	DeleteGoal(context.Context, data.DeleteGoalParams) error
}
