package budget

import (
	"time"

	"samuellando.com/YNAFB/internal/account"
	"samuellando.com/YNAFB/internal/category"
	"samuellando.com/YNAFB/internal/goal"
)

type Budget struct {
	Accounts []*account.Account
	Goals []*goal.Goal
	Months MonthlyBudget
}

type MonthlyBudget struct {
	Date time.Time
	Allocations []Allocation
}

type Allocation struct {
	category *category.Category
	amount int
}
