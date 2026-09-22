package domain

import (
	"fmt"

	"samuellando.com/YNAFB/data"
)

type TrxLine struct {
	service            *TrxService
	row                data.TrxLine
	category           *Category
	destinationAccount *Account
}

func (l *TrxLine) ID() int {
	return int(l.row.ID)
}

func (l *TrxLine) IsIncome() bool {
	return l.row.Income
}

func (l *TrxLine) Category() (*Category, error) {
	if l.row.CategoryID.Valid {
		return l.category, nil
	}
	return nil, fmt.Errorf("Transaction line is not a categorization")
}

func (l *TrxLine) DestinationAccount() (*Account, error) {
	if l.row.DestAccountID.Valid {
		return l.destinationAccount, nil
	}
	return nil, fmt.Errorf("Transaction line is not a transfer")
}

func (l *TrxLine) Inflow() int {
	return int(l.row.Inflow)
}

func (l *TrxLine) Outflow() int {
	return int(l.row.Outflow)
}

