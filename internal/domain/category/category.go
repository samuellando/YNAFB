package category

import (
	"samuellando.com/YNAFB/data"
)

type Category struct {
	row data.Category
	group *Group
}

type Group struct {
	row data.CategoryGroup
}

func GroupFromRow(row data.CategoryGroup) *Group {
	return &Group{
		row: row,
	}
}

func FromRow(row data.Category, group *Group) *Category {
	return &Category{
		row: row,
		group: group,
	}
}
