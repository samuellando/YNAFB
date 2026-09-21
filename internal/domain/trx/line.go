package trx

import (
	"fmt"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/domain/account"
	"samuellando.com/YNAFB/internal/domain/category"
)

type Line struct {
	service            *Service
	row                data.TrxLine
	category           *category.Category
	destinationAccount *account.Account
}

func (l *Line) IsIncome() bool {
	return l.row.Income
}

func (l *Line) Category() (*category.Category, error) {
	if l.row.CategoryID.Valid {
		return l.category, nil
	}
	return nil, fmt.Errorf("Transaction line is not a categorization")
}

func (l *Line) DestinationAccount() (*account.Account, error) {
	if l.row.DestAccountID.Valid {
		return l.destinationAccount, nil
	}
	return nil, fmt.Errorf("Transaction line is not a transfer")
}

func (l *Line) Inflow() int {
	return int(l.row.Inflow)
}

func (l *Line) Outflow() int {
	return int(l.row.Outflow)
}
