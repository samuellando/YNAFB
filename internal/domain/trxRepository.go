package domain

import (
	"context"
	"samuellando.com/YNAFB/data"
)

type TrxRepository interface {
	// Trx Operations
	ListTrxsAndLines(context.Context, data.ListTrxsAndLinesParams) ([]data.ListTrxsAndLinesRow, error)
	CreateTrx(context.Context, data.CreateTrxParams) (data.Trx, error)
}
