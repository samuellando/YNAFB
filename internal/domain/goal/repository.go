package goal

import (
	"context"
	"samuellando.com/YNAFB/data"
)

type Repository interface {
	ListGoals(context.Context, data.ListGoalsParams) ([]data.Goal, error)
}
