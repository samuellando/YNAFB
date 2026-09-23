package domain

import (
	"context"
	"samuellando.com/YNAFB/data"
)

type TrxRepository interface {
	// Trx Operations
	ListTrxsAndLines(context.Context, data.ListTrxsAndLinesParams) ([]data.ListTrxsAndLinesRow, error)
	GetTrxAndLines(context.Context, data.GetTrxAndLinesParams) ([]data.GetTrxAndLinesRow, error)
	CreateTrx(context.Context, data.CreateTrxParams) (data.Trx, error)
	UpdateTrx(context.Context, data.UpdateTrxParams) (data.Trx, error)
	DeleteTrx(context.Context, data.DeleteTrxParams) error
	// TrxLine Operations
	GetTrxLine(context.Context, data.GetTrxLineParams) (data.GetTrxLineRow, error)
	CreateTrxLine(context.Context, data.CreateTrxLineParams) (data.TrxLine, error)
	UpdateTrxLine(context.Context, data.UpdateTrxLineParams) (data.TrxLine, error)
	DeleteTrxLine(context.Context, data.DeleteTrxLineParams) error
}
