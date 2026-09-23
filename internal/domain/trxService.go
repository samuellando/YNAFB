package domain

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/db/types"
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

func (s *TrxService) WithRepo(repo TrxRepository) *TrxService {
	cp := *s
	cp.repo = repo
	return &cp
}

func (s *TrxService) SetBudgetService(budgetService *BudgetService) {
	s.budgetService = budgetService
}

func nullInt64FromAccount(a *Account) sql.NullInt64 {
	if a == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(a.ID()), Valid: true}
}

func nullInt64FromCategory(c *Category) sql.NullInt64 {
	if c == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(c.ID()), Valid: true}
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

func (s *TrxService) Create(ctx context.Context, account *Account, payee *Payee, date time.Time, outflow, inflow int, note string) (*Trx, error) {
	defer cache.InvalidateResults(ctx)
	row, err := s.repo.CreateTrx(ctx, data.CreateTrxParams{
		LoginID:      int64(account.Budget().LoginID()),
		BudgetID:     int64(account.Budget().ID()),
		Date:         types.UnixTime{Time: date},
		AccountID:    int64(account.ID()),
		PayeeID:      int64(payee.ID()),
		TotalOutflow: int64(outflow),
		TotalInflow:  int64(inflow),
		Note:         note,
	})
	if err != nil {
		return nil, err
	}
	trx := &Trx{
		service: s,
		row:     row,
		budget:  account.Budget(),
		account: account,
		payee:   payee,
	}
	cache.Store(ctx, row.ID, trx)
	return trx, nil
}

func (s *TrxService) Get(ctx context.Context, loginID, budgetID, id int) (*Trx, error) {
	if v, ok := cache.Get[*Trx](ctx, int64(id)); ok {
		return v, nil
	}
	rows, err := s.repo.GetTrxAndLines(ctx, data.GetTrxAndLinesParams{
		LoginID:  int64(loginID),
		BudgetID: int64(budgetID),
		ID:       int64(id),
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	_, trx := s.loadTrx(ctx, asListRows(rows))
	return trx, nil
}

// asListRows maps GetTrxAndLinesRow to its field-identical ListTrxsAndLinesRow
// so Get can rely on loadTrx (Go forbids direct conversion: both are named
// types).
func asListRows(rows []data.GetTrxAndLinesRow) []data.ListTrxsAndLinesRow {
	out := make([]data.ListTrxsAndLinesRow, len(rows))
	for i, r := range rows {
		out[i] = data.ListTrxsAndLinesRow{
			Trx:               r.Trx,
			BudgetID:          r.BudgetID,
			BudgetName:        r.BudgetName,
			LoginID:           r.LoginID,
			AccountID:         r.AccountID,
			AccountName:       r.AccountName,
			PayeeID:           r.PayeeID,
			PayeeName:         r.PayeeName,
			LineID:            r.LineID,
			LineIncome:        r.LineIncome,
			LineInflow:        r.LineInflow,
			LineOutlfow:       r.LineOutlfow,
			CategoryID:        r.CategoryID,
			CategoryName:      r.CategoryName,
			CategoryGroupID:   r.CategoryGroupID,
			CategoryGroupName: r.CategoryGroupName,
			DestAccountID:     r.DestAccountID,
			DestAccountName:   r.DestAccountName,
			Reconciled:        r.Reconciled,
		}
	}
	return out
}

func (s *TrxService) CreateLine(ctx context.Context, loginID, budgetID, trxID int, dest *Account, category *Category, income bool, outflow, inflow int) (*TrxLine, error) {
	defer cache.InvalidateResults(ctx)
	row, err := s.repo.CreateTrxLine(ctx, data.CreateTrxLineParams{
		TrxID:         int64(trxID),
		DestAccountID: nullInt64FromAccount(dest),
		CategoryID:    nullInt64FromCategory(category),
		Income:        income,
		Outflow:       int64(outflow),
		Inflow:        int64(inflow),
		BudgetID:      int64(budgetID),
		LoginID:       int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	line := &TrxLine{
		service:            s,
		row:                row,
		category:           category,
		destinationAccount: dest,
	}
	cache.Store(ctx, line.row.ID, line)
	return line, nil
}

func (s *TrxService) GetLine(ctx context.Context, loginID, budgetID, id int) (*TrxLine, error) {
	if v, ok := cache.Get[*TrxLine](ctx, int64(id)); ok {
		return v, nil
	}
	row, err := s.repo.GetTrxLine(ctx, data.GetTrxLineParams{
		LoginID:  int64(loginID),
		BudgetID: int64(budgetID),
		ID:       int64(id),
	})
	if err != nil {
		return nil, err
	}
	budget := s.budgetService.fromRow(ctx, data.Budget{
		ID:      row.TrxLine.BudgetID,
		LoginID: int64(loginID),
		Name:    row.BudgetName,
	})
	line := s.loadTrxLine(ctx, budget, row.TrxLine, row.DestAccountName, row.CategoryName, row.CategoryGroupName, row.CategoryGroupID)
	cache.Store(ctx, line.row.ID, line)
	return line, nil
}

// loadTrxLine builds a TrxLine entity with its category/destination refs,
// shared by loadTrx and GetLine. Callers own instance caching.
func (s *TrxService) loadTrxLine(ctx context.Context, budget *Budget, row data.TrxLine, destName, catName, groupName sql.NullString, groupID sql.NullInt64) *TrxLine {
	line := &TrxLine{
		service: s,
		row:     row,
	}
	if row.CategoryID.Valid {
		var group *Group
		if groupID.Valid {
			group = s.categoryService.groupFromRow(ctx, data.CategoryGroup{
				ID:       groupID.Int64,
				BudgetID: row.BudgetID,
				Name:     groupName.String,
			})
		}
		line.category = s.categoryService.fromRow(ctx, data.Category{
			ID:              row.CategoryID.Int64,
			BudgetID:        row.BudgetID,
			Name:            catName.String,
			CategoryGroupID: groupID,
		}, group)
	}
	if row.DestAccountID.Valid {
		line.destinationAccount = s.accountService.fromRow(ctx, data.Account{
			ID:       row.DestAccountID.Int64,
			BudgetID: row.BudgetID,
			Name:     destName.String,
		}, budget)
	}
	return line
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
			line := s.loadTrxLine(ctx, budget, data.TrxLine{
				ID:            row.LineID.Int64,
				BudgetID:      row.Trx.BudgetID,
				TrxID:         row.Trx.ID,
				DestAccountID: row.DestAccountID,
				CategoryID:    row.CategoryID,
				Income:        row.LineIncome.Bool,
				Outflow:       row.LineOutlfow.Int64,
				Inflow:        row.LineInflow.Int64,
			}, row.DestAccountName, row.CategoryName, row.CategoryGroupName, row.CategoryGroupID)
			trx.lines = append(trx.lines, line)
		}
	}
	cache.Store(ctx, trx.row.ID, trx)
	return len(rows), &trx
}
