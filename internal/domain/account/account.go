package account

import "samuellando.com/YNAFB/data"

type Account struct {
	row data.Account
}

func FromRow(row data.Account) *Account {
	return &Account{row: row}
}
