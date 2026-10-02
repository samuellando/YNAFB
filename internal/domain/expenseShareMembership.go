package domain

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"samuellando.com/YNAFB/internal/cache"
	"samuellando.com/YNAFB/internal/data"
)

type ExpenseShareMembership struct {
	row    data.BudgetExpenseShare
	budget *Budget
	share  *ExpenseShare
}

type MemberBalance struct {
	Budget      *Budget
	DisplayName string
	Balance     int
}

func expenseShareMembershipFromRow(ctx context.Context, row data.BudgetExpenseShare, budget *Budget, share *ExpenseShare) *ExpenseShareMembership {
	membership, _ := cache.Get(ctx, row.ID, func() (*ExpenseShareMembership, error) {
		return &ExpenseShareMembership{
			row:    row,
			budget: budget,
			share:  share,
		}, nil
	})
	return membership
}

// Hydrate a membership from a wide member row (bes + es + budget in one trip).
func membershipFromMemberRow(ctx context.Context, bes data.BudgetExpenseShare, es data.ExpenseShare, b data.Budget, service *DomainService) *ExpenseShareMembership {
	budget := budgetFromRow(ctx, b, service)
	share := expenseShareFromRow(ctx, es, service)
	return expenseShareMembershipFromRow(ctx, bes, budget, share)
}

func (b *Budget) ListExpenseShares(ctx context.Context) ([]*ExpenseShareMembership, error) {
	return cache.Result(ctx, fmt.Sprintf("budgetExpenseShareList-%d-%d", b.LoginID(), b.ID()), func() ([]*ExpenseShareMembership, error) {
		rows, err := b.service.repo.ListMembershipsByBudget(ctx, data.ListMembershipsByBudgetParams{
			LoginID:  int64(b.LoginID()),
			BudgetID: int64(b.ID()),
		})
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return nil, nil
		}
		memberships := make([]*ExpenseShareMembership, len(rows))
		for i, row := range rows {
			memberships[i] = membershipFromMemberRow(ctx, row.BudgetExpenseShare, row.ExpenseShare, row.Budget, b.service)
		}
		return memberships, nil
	})
}

func (b *Budget) CreateExpenseShare(ctx context.Context, name, displayName string, defaultName *string) (*ExpenseShareMembership, error) {
	defer cache.InvalidateResults(ctx)
	resolved := name
	if defaultName != nil {
		resolved = *defaultName
	}
	shareRow, err := b.service.repo.CreateExpenseShare(ctx, data.CreateExpenseShareParams{
		DefaultName: resolved,
	})
	if err != nil {
		return nil, err
	}
	share := expenseShareFromRow(ctx, shareRow, b.service)
	row, err := b.service.repo.CreateMembership(ctx, data.CreateMembershipParams{
		Name:           name,
		DisplayName:    displayName,
		ExpenseShareID: shareRow.ID,
		BudgetID:       int64(b.ID()),
		LoginID:        int64(b.LoginID()),
	})
	if err != nil {
		return nil, err
	}
	return expenseShareMembershipFromRow(ctx, row, b, share), nil
}

func (b *Budget) JoinExpenseShare(ctx context.Context, code, name, displayName string) (*ExpenseShareMembership, error) {
	defer cache.InvalidateResults(ctx)
	found, err := b.service.repo.GetExpenseShareByCode(ctx, data.GetExpenseShareByCodeParams{
		Code: code,
	})
	if err != nil {
		return nil, err
	}
	if found.Expires.Time.Before(time.Now()) {
		return nil, fmt.Errorf("invite code expired")
	}
	share := expenseShareFromRow(ctx, found.ExpenseShare, b.service)
	row, err := b.service.repo.CreateMembership(ctx, data.CreateMembershipParams{
		Name:           name,
		DisplayName:    displayName,
		ExpenseShareID: found.ExpenseShare.ID,
		BudgetID:       int64(b.ID()),
		LoginID:        int64(b.LoginID()),
	})
	if err != nil {
		return nil, err
	}
	return expenseShareMembershipFromRow(ctx, row, b, share), nil
}

func (b *Budget) GetExpenseShare(ctx context.Context, shareID int) (*ExpenseShareMembership, error) {
	return cache.Result(ctx, fmt.Sprintf("budgetExpenseShare-%d-%d-%d", b.LoginID(), b.ID(), shareID), func() (*ExpenseShareMembership, error) {
		row, err := b.service.repo.GetMembershipByShare(ctx, data.GetMembershipByShareParams{
			LoginID:        int64(b.LoginID()),
			BudgetID:       int64(b.ID()),
			ExpenseShareID: int64(shareID),
		})
		if err != nil {
			return nil, err
		}
		return membershipFromMemberRow(ctx, row.BudgetExpenseShare, row.ExpenseShare, row.Budget, b.service), nil
	})
}

func (m *ExpenseShareMembership) ID() int {
	return int(m.row.ID)
}

func (m *ExpenseShareMembership) Name() string {
	return m.row.Name
}

func (m *ExpenseShareMembership) DisplayName() string {
	return m.row.DisplayName
}

func (m *ExpenseShareMembership) Budget() *Budget {
	return m.budget
}

func (m *ExpenseShareMembership) Share() *ExpenseShare {
	return m.share
}

func (m *ExpenseShareMembership) Members(ctx context.Context) ([]*ExpenseShareMembership, error) {
	return cache.Result(ctx, fmt.Sprintf("shareMembers-%d-%d-%d", m.budget.LoginID(), m.budget.ID(), m.share.ID()), func() ([]*ExpenseShareMembership, error) {
		rows, err := m.budget.service.repo.ListShareMembers(ctx, data.ListShareMembersParams{
			ExpenseShareID: int64(m.share.ID()),
			BudgetID:       int64(m.budget.ID()),
			LoginID:        int64(m.budget.LoginID()),
		})
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return nil, nil
		}
		members := make([]*ExpenseShareMembership, len(rows))
		for i, row := range rows {
			members[i] = membershipFromMemberRow(ctx, row.BudgetExpenseShare, row.ExpenseShare, row.Budget, m.budget.service)
		}
		return members, nil
	})
}

func (m *ExpenseShareMembership) Balances(ctx context.Context) (int, []MemberBalance, error) {
	members, err := m.Members(ctx)
	if err != nil {
		return 0, nil, err
	}
	transactions, err := m.ListTransactions(ctx)
	if err != nil {
		return 0, nil, err
	}
	byBudget := make(map[int]*MemberBalance, len(members))
	order := make([]int, 0, len(members))
	for _, member := range members {
		if member.Budget().ID() == m.budget.ID() {
			continue
		}
		byBudget[member.Budget().ID()] = &MemberBalance{
			Budget:      member.Budget(),
			DisplayName: member.DisplayName(),
		}
		order = append(order, member.Budget().ID())
	}
	total := 0
	add := func(budgetID int, amount int) {
		total += amount
		if pair, ok := byBudget[budgetID]; ok {
			pair.Balance += amount
		}
	}
	for _, trx := range transactions {
		sourceID := trx.SourceBudget().ID()
		for _, line := range trx.SourceTrx().Lines() {
			share, err := line.Share()
			if err != nil || share.ID() != m.share.ID() {
				continue
			}
			amount := line.Outflow() - line.Inflow()
			if split, err := line.SplitBudget(); err == nil {
				otherID := split.Budget().ID()
				switch {
				case sourceID == m.budget.ID() && otherID != m.budget.ID():
					add(otherID, amount)
				case otherID == m.budget.ID() && sourceID != m.budget.ID():
					add(sourceID, -amount)
				}
			}
			if dest, err := line.DestBudget(); err == nil {
				otherID := dest.Budget().ID()
				switch {
				case sourceID == m.budget.ID() && otherID != m.budget.ID():
					add(otherID, -amount)
				case otherID == m.budget.ID() && sourceID != m.budget.ID():
					add(sourceID, amount)
				}
			}
		}
	}
	balances := make([]MemberBalance, 0, len(order))
	for _, budgetID := range order {
		if pair := byBudget[budgetID]; pair.Balance != 0 {
			balances = append(balances, *pair)
		}
	}
	return total, balances, nil
}

func (m *ExpenseShareMembership) Update(ctx context.Context, name, displayName *string) error {
	defer cache.InvalidateResults(ctx)
	resolvedName := m.row.Name
	if name != nil {
		resolvedName = *name
	}
	resolvedDisplay := m.row.DisplayName
	if displayName != nil {
		resolvedDisplay = *displayName
	}
	row, err := m.budget.service.repo.UpdateMembership(ctx, data.UpdateMembershipParams{
		Name:           resolvedName,
		DisplayName:    resolvedDisplay,
		ExpenseShareID: int64(m.share.ID()),
		BudgetID:       int64(m.budget.ID()),
		LoginID:        int64(m.budget.LoginID()),
	})
	if err != nil {
		return err
	}
	m.row = row
	return nil
}

func (m *ExpenseShareMembership) Leave(ctx context.Context) error {
	defer cache.InvalidateResults(ctx)
	defer cache.Delete[*ExpenseShareMembership](ctx, int64(m.ID()))
	return m.budget.service.repo.DeleteMembership(ctx, data.DeleteMembershipParams{
		ExpenseShareID: int64(m.share.ID()),
		BudgetID:       int64(m.budget.ID()),
		LoginID:        int64(m.budget.LoginID()),
	})
}

func (m *ExpenseShareMembership) ListTransactions(ctx context.Context) ([]*ExpenseShareTransaction, error) {
	return cache.Result(ctx, fmt.Sprintf("shareTransactions-%d-%d-%d", m.budget.LoginID(), m.budget.ID(), m.share.ID()), func() ([]*ExpenseShareTransaction, error) {
		rows, err := m.budget.service.repo.ListShareTransactions(ctx, data.ListShareTransactionsParams{
			ExpenseShareID: int64(m.share.ID()),
			BudgetID:       int64(m.budget.ID()),
			LoginID:        int64(m.budget.LoginID()),
		})
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return nil, nil
		}
		return shareTransactionsFromListRows(ctx, m, rows)
	})
}

func (m *ExpenseShareMembership) GetTransaction(ctx context.Context, trxID int) (*ExpenseShareTransaction, error) {
	rows, err := m.budget.service.repo.GetShareTransaction(ctx, data.GetShareTransactionParams{
		ExpenseShareID: int64(m.share.ID()),
		TrxID:          int64(trxID),
		BudgetID:       int64(m.budget.ID()),
		LoginID:        int64(m.budget.LoginID()),
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	listRows := make([]data.ListShareTransactionsRow, len(rows))
	for i, row := range rows {
		listRows[i] = data.ListShareTransactionsRow(row)
	}
	transactions, err := shareTransactionsFromListRows(ctx, m, listRows)
	if err != nil {
		return nil, err
	}
	if len(transactions) == 0 {
		return nil, sql.ErrNoRows
	}
	return transactions[0], nil
}

// Build share transaction views from wide share rows. Rows for one source
// transaction are contiguous (ORDER BY date, id); each group hydrates through
// trxFromRows with its own source budget, like listTransactions does.
func shareTransactionsFromListRows(ctx context.Context, m *ExpenseShareMembership, rows []data.ListShareTransactionsRow) ([]*ExpenseShareTransaction, error) {
	transactions := make([]*ExpenseShareTransaction, 0)
	n := 0
	for n < len(rows) {
		sourceBudget := budgetFromRow(ctx, data.Budget{
			ID:      rows[n].SourceBudgetID,
			LoginID: rows[n].SourceBudgetLoginID,
			Name:    rows[n].SourceBudgetName,
		}, m.budget.service)
		displayName := rows[n].SourceMemberDisplayName
		end := n + 1
		for end < len(rows) && rows[end].Trx.ID == rows[n].Trx.ID {
			end++
		}
		listRows := make([]data.ListTrxsAndLinesRow, end-n)
		for i, row := range rows[n:end] {
			listRows[i] = shareListRow(row)
		}
		_, trx := trxFromRows(ctx, listRows, sourceBudget)
		if trx == nil {
			return nil, sql.ErrNoRows
		}
		transactions = append(transactions, expenseShareTransactionFromSource(m, sourceBudget, trx, displayName))
		n = end
	}
	return transactions, nil
}

func shareListRow(row data.ListShareTransactionsRow) data.ListTrxsAndLinesRow {
	return data.ListTrxsAndLinesRow{
		Trx:                                row.Trx,
		AccountName:                        row.AccountName,
		PayeeName:                          row.PayeeName,
		LineID:                             row.LineID,
		LineIncome:                         row.LineIncome,
		LineInflow:                         row.LineInflow,
		LineOutflow:                        row.LineOutflow,
		CategoryID:                         row.CategoryID,
		CategoryName:                       row.CategoryName,
		CategoryGroupID:                    row.CategoryGroupID,
		CategoryGroupName:                  row.CategoryGroupName,
		DestAccountID:                      row.DestAccountID,
		DestAccountName:                    row.DestAccountName,
		ExpenseShareDefaultName:            row.ExpenseShareDefaultName,
		ExpenseShareID:                     row.ExpenseShareID,
		SplitBudgetExpenseShareID:          row.SplitBudgetExpenseShareID,
		DestBudgetExpenseShareID:           row.DestBudgetExpenseShareID,
		SplitBudgetExpenseShareName:        row.SplitBudgetExpenseShareName,
		SplitBudgetExpenseShareDisplayName: row.SplitBudgetExpenseShareDisplayName,
		SplitBudgetID:                      row.SplitBudgetID,
		SplitBudgetName:                    row.SplitBudgetName,
		SplitBudgetLoginID:                 row.SplitBudgetLoginID,
		DestBudgetExpenseShareName:         row.DestBudgetExpenseShareName,
		DestBudgetExpenseShareDisplayName:  row.DestBudgetExpenseShareDisplayName,
		DestBudgetID:                       row.DestBudgetID,
		DestBudgetName:                     row.DestBudgetName,
		DestBudgetLoginID:                  row.DestBudgetLoginID,
		Reconciled:                         row.Reconciled,
	}
}
