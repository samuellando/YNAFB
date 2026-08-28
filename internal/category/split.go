package category

import (
	"fmt"
	"slices"
)

type CategorySplit struct {
	Percent uint8
	Category *Category
}

type CategorySplitProfile struct {
	splits []CategorySplit
}

func CreateCategorySplitProfile(splits []CategorySplit) (*CategorySplitProfile, error) {
	var total uint8 = 0
	for _, split := range splits {
		total += split.Percent
	}
	if total != 100 {
		return nil, fmt.Errorf("Split Percentages must equal 100")
	}
	return &CategorySplitProfile{splits: slices.Clone(splits)}, nil
}

func (csp CategorySplitProfile) Splits() []CategorySplit {
	return slices.Clone(csp.splits)
}
