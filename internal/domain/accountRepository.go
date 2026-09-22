package domain

import (
	"context"

	"samuellando.com/YNAFB/data"
)

type AccountRepository interface {
   ListAccounts(context.Context, data.ListAccountsParams) ([]data.ListAccountsRow, error)
   GetAccount(context.Context, data.GetAccountParams) (data.GetAccountRow, error)
   CreateAccount(context.Context, data.CreateAccountParams) (data.CreateAccountRow, error)
   DeleteAccount(context.Context, data.DeleteAccountParams) error 
   UpdateAccount(context.Context, data.UpdateAccountParams) (data.Account, error)
   ReconcileAccountTransactions(context.Context, data.ReconcileAccountTransactionsParams) (int64, error)
}
