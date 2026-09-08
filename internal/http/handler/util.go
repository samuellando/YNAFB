package handler

import (
	"fmt"
	"net/http"
)

func getLoginID(req *http.Request) (int64, error) {
	idA := req.Context().Value("loginID")
	if idA == nil {
		return 0, fmt.Errorf("No login ID")
	}
	if id, ok := idA.(int64); ok {
		return id, nil
	} else {
		return 0, fmt.Errorf("Id not int64")
	}
}
