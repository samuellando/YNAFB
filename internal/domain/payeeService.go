package domain

import (
)

type PayeeService struct {
	repo PayeeRepository
}

func NewPayeeService(repo PayeeRepository) *PayeeService {
	return &PayeeService{repo: repo}
}
