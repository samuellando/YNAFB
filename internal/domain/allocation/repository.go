package allocation

import (
	"context"
	"samuellando.com/YNAFB/data"
)

type Repository interface {
	ListAllocations(context.Context, data.ListAllocationsParams) ([]data.Allocation, error)
}
