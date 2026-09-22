package domain

import (
	"context"
	"samuellando.com/YNAFB/data"
)

type AloocationRepository interface {
	ListAllocations(context.Context, data.ListAllocationsParams) ([]data.Allocation, error)
	SetAllocation(context.Context, data.SetAllocationParams) (data.Allocation, error)
}
