package domain

import (
	"context"
	"samuellando.com/YNAFB/data"
)

type GoalRepository interface {
	ListGoals(context.Context, data.ListGoalsParams) ([]data.Goal, error)
}
