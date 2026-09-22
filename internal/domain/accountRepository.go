package domain

import (
	"context"

	"samuellando.com/YNAFB/data"
)

type AccountRepository interface {
   ListAccounts(context.Context, data.ListAccountsParams) ([]data.Account, error)
   GetAccount(context.Context, data.GetAccountParams) (data.Account, error)
   CreateAccount(context.Context, data.CreateAccountParams) (data.Account, error)
   DeleteAccount(context.Context, data.DeleteAccountParams) error 
   UpdateAccount(context.Context, data.UpdateAccountParams) (data.Account, error)
}
