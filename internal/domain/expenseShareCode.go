package domain

import (
	"context"
	"time"

	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/data"
)

type ExpenseShareCode struct {
	row   data.ExpenseShareCode
	share *ExpenseShare
}

func expenseShareCodeFromRow(ctx context.Context, row data.ExpenseShareCode, share *ExpenseShare) *ExpenseShareCode {
	code, _ := cache.Get(ctx, row.ID, func() (*ExpenseShareCode, error) {
		return &ExpenseShareCode{
			row:   row,
			share: share,
		}, nil
	})
	return code
}

// Get the invite code.
func (c *ExpenseShareCode) Code() string {
	return c.row.Code
}

// Get when the code expires (server-decided).
func (c *ExpenseShareCode) Expires() time.Time {
	return c.row.Expires.Time
}

// Get the share this code was minted for.
func (c *ExpenseShareCode) Share() *ExpenseShare {
	return c.share
}
