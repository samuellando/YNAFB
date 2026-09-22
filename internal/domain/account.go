package domain 

import (
	"context"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type Account struct {
	service *AccountService
	row data.Account
}

func (s *AccountService) FromRow(ctx context.Context, row data.Account) *Account {
	if cached, ok := cache.Get[*Account](ctx, row.ID); ok {
		return cached
	}
	account := &Account{
		service: s,
		row: row,
	}
	cache.Store(ctx, row.ID, account)
	return account
}
