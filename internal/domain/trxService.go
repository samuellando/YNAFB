package domain

import (
	"context"
	"fmt"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
)

type TrxService struct {
	repo            TrxRepository
	categoryService *CategoryService
	payeeService    *PayeeService
	accountService  *AccountService
	budgetService   *BudgetService
}

func NewTrxService(repo TrxRepository, catcategoryService *CategoryService, payeeService *PayeeService, accountService *AccountService, budgetSerivce *BudgetService) *TrxService {
	return &TrxService{
		repo:            repo,
		categoryService: catcategoryService,
		payeeService:    payeeService,
		accountService:  accountService,
		budgetService:   budgetSerivce,
	}
}

func (s *TrxService) list(ctx context.Context, loginID, budgetID int) ([]*Trx, error) {
	return cache.Result(ctx, fmt.Sprintf("trxServiceList-%d-%d", loginID, budgetID), func() ([]*Trx, error) {
		rows, err := s.repo.ListTrxsAndLines(ctx, data.ListTrxsAndLinesParams{
			LoginID:  int64(loginID),
			BudgetID: int64(budgetID),
		})
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return nil, nil
		}
		trxs := make([]*Trx, 0)
		n := 0
		for n < len(rows) {
			consumed, trx := s.loadTrx(ctx, rows[n:])
			trxs = append(trxs, trx)
			n += consumed
		}
		return trxs, nil
	})
}

func (s *TrxService) loadTrx(ctx context.Context, rows []data.ListTrxsAndLinesRow) (int, *Trx) {
	if len(rows) == 0 {
		return 0, nil
	}
	if trx, ok := cache.Get[*Trx](ctx, rows[0].Trx.ID); ok {
		for i, row := range rows {
			if row.Trx.ID != trx.row.ID {
				return i, trx
			}
		}
	}
	budget := s.budgetService.fromRow(ctx, data.Budget{
		ID:      rows[0].BudgetID,
		LoginID: rows[0].LoginID,
		Name:    rows[0].BudgetName,
	})
	account := s.accountService.fromRow(ctx, data.Account{
		ID:   rows[0].AccountID,
		Name: rows[0].AccountName,
	}, budget)
	trx := Trx{
		service: s,
		row:     rows[0].Trx,
		budget:  budget,
		account: account,
		payee: s.payeeService.fromRow(ctx,
			data.Payee{
				ID:       rows[0].PayeeID,
				BudgetID: rows[0].Trx.BudgetID,
				Name:     rows[0].PayeeName,
			},
		),
		reconciled: rows[0].Reconciled,
	}
	for i, row := range rows {
		if row.Trx.ID != trx.row.ID {
			return i, &trx
		}
		if row.LineID.Valid {
			line := TrxLine{
				service: s,
				row: data.TrxLine{
					ID:            row.LineID.Int64,
					BudgetID:      row.Trx.BudgetID,
					TrxID:         row.Trx.ID,
					DestAccountID: row.DestAccountID,
					CategoryID:    row.CategoryID,
					Income:        row.LineIncome.Bool,
					Outflow:       row.LineOutlfow.Int64,
					Inflow:        row.LineInflow.Int64,
				},
			}
			if row.CategoryID.Valid {
				var group *Group
				if row.CategoryGroupID.Valid {
					group = s.categoryService.groupFromRow(ctx, data.CategoryGroup{
						ID:       row.CategoryGroupID.Int64,
						BudgetID: row.Trx.BudgetID,
						Name:     row.CategoryGroupName.String,
					})
				}
				line.category = s.categoryService.fromRow(ctx, data.Category{
					ID:              row.CategoryID.Int64,
					BudgetID:        row.Trx.BudgetID,
					Name:            row.CategoryName.String,
					CategoryGroupID: row.CategoryGroupID,
				},
					group,
				)
			}
			if row.DestAccountID.Valid {
				line.destinationAccount = s.accountService.fromRow(ctx, data.Account{
					ID:       row.DestAccountID.Int64,
					BudgetID: row.Trx.BudgetID,
					Name:     row.DestAccountName.String,
				}, budget)
			}
			trx.lines = append(trx.lines, &line)
		}
	}
	cache.Store(ctx, trx.row.ID, trx)
	return len(rows), &trx
}
