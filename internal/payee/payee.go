package payee

import (
	"samuellando.com/YNAFB/internal/category"
)

type Payee struct {
	Name            string
	DefaultCategorySplit *category.CategorySplitProfile
}

