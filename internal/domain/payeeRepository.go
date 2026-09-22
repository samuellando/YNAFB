package domain

import (
	"context"

	"samuellando.com/YNAFB/data"
)

type PayeeRepository interface {
	ListPayees(context.Context, data.ListPayeesParams) ([]data.Payee, error)
	GetPayee(context.Context, data.GetPayeeParams) (data.Payee, error)
	GetPayeeByName(context.Context, data.GetPayeeByNameParams) (data.Payee, error)
	CreatePayee(context.Context, data.CreatePayeeParams) (data.Payee, error)
	UpdatePayee(context.Context, data.UpdatePayeeParams) (data.Payee, error)
	DeletePayee(context.Context, data.DeletePayeeParams) error
	ListPayeeDefaultLinesByPayee(context.Context, data.ListPayeeDefaultLinesByPayeeParams) ([]data.ListPayeeDefaultLinesByPayeeRow, error)
	GetPayeeDefaultLine(context.Context, data.GetPayeeDefaultLineParams) (data.PayeeDefaultLine, error)
	CreatePayeeDefaultLine(context.Context, data.CreatePayeeDefaultLineParams) (data.PayeeDefaultLine, error)
	UpdatePayeeDefaultLine(context.Context, data.UpdatePayeeDefaultLineParams) (data.PayeeDefaultLine, error)
	DeletePayeeDefaultLine(context.Context, data.DeletePayeeDefaultLineParams) error
}
