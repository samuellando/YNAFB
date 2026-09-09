package api

import (
	"context"
	"fmt"
)

func getLoginID(ctx context.Context) (int64, error) {
	idA := ctx.Value("loginID")
	if idA == nil {
		return 0, fmt.Errorf("No login ID")
	}
	if id, ok := idA.(int64); ok {
		return id, nil
	} else {
		return 0, fmt.Errorf("Id not int64")
	}
}
