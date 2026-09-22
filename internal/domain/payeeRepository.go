package domain

import (
	"context"

	"samuellando.com/YNAFB/data"
)

type PayeeRepository interface {
	GetPayeeByName(context.Context, data.GetPayeeByNameParams) (data.Payee, error)
	CreatePayee(context.Context, data.CreatePayeeParams) (data.Payee, error)
}
