package account

import (
	"samuellando.com/YNAFB/internal/transaction"
)

type AccountType string

const (
	Asset AccountType = "ASSET"
	Debt AccountType = "DEBT"
)

type Account struct {
	AccountType AccountType
	Name string
	Transactions []*transaction.Transaction
}
