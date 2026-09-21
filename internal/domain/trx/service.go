package trx

import (
	"context"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/domain/account"
	"samuellando.com/YNAFB/internal/domain/category"
	"samuellando.com/YNAFB/internal/domain/payee"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, loginID, budgetID int) ([]*Trx, error) {
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
	for ; n < len(rows); {
		consumed, trx := loadTrx(s, rows[n:])
		trxs = append(trxs, trx)
		n += consumed
	}
	return trxs, nil
}

func loadTrx(s *Service, rows []data.ListTrxsAndLinesRow) (int, *Trx) {
	if len(rows) == 0 {
		return 0, nil
	}
	trx := Trx{
		service: s,
		row:     rows[0].Trx,
		payee: *payee.FromRow(
			data.Payee{
				ID:       rows[0].PayeeID,
				BudgetID: rows[0].Trx.BudgetID,
				Name:     rows[0].PayeeName,
			},
		),
	}
	for i, row := range rows {
		if row.Trx.ID != trx.row.ID {
			return i, &trx
		}
		if row.LineID.Valid {
			line := Line{
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
				var group *category.Group
				if row.CategoryGroupID.Valid {
					group = category.GroupFromRow(data.CategoryGroup{
						ID:       row.CategoryGroupID.Int64,
						BudgetID: row.Trx.BudgetID,
						Name:     row.CategoryGroupName.String,
					})
				}
				line.category = category.FromRow(data.Category{
					ID:              row.CategoryID.Int64,
					BudgetID:        row.Trx.BudgetID,
					Name:            row.CategoryName.String,
					CategoryGroupID: row.CategoryGroupID,
				},
					group,
				)
			}
			if row.DestAccountID.Valid {
				line.destinationAccount = account.FromRow(data.Account{
					ID:       row.DestAccountID.Int64,
					BudgetID: row.Trx.BudgetID,
					Name:     row.DestAccountName.String,
				})
			}
			trx.lines = append(trx.lines, &line)
		}
	}
	return len(rows), &trx
}
