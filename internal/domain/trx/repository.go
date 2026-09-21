package trx

import (
	"context"
	"samuellando.com/YNAFB/data"
)

type Repository interface {
	// Trx Operations
	ListTrxsAndLines(context.Context, data.ListTrxsAndLinesParams) ([]data.ListTrxsAndLinesRow, error)
}
