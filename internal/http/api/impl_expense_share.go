package api

import (
	"context"
	"database/sql"
	"time"
	"fmt"
	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/config"
	"strconv"
	"github.com/google/uuid"
)

func (s ApiServer) GetBudgetBudgetIdExpenseShare(ctx context.Context, request GetBudgetBudgetIdExpenseShareRequestObject) (GetBudgetBudgetIdExpenseShareResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetString := request.BudgetId
	budgetID, err := strconv.Atoi(budgetString)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	// Query the database
	shares, err := s.queries.ListBudgetExpenseShares(ctx, data.ListBudgetExpenseSharesParams{
		LoginID:  loginID,
		BudgetID: int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	// Send the response
	resp := make(GetBudgetBudgetIdExpenseShare200JSONResponse, 0)
	for _, share := range shares {
		resp = append(resp, BudgetExpenseShare{
			DisplayName: share.DisplayName,
			Name:        share.Name,
			Id:          int(share.ExpenseShareID),
		})
	}
	return resp, nil
}

func (s ApiServer) PostBudgetBudgetIdExpenseShare(ctx context.Context, request PostBudgetBudgetIdExpenseShareRequestObject) (PostBudgetBudgetIdExpenseShareResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetString := request.BudgetId
	budgetID, err := strconv.Atoi(budgetString)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	// Query the database
	var expense_share data.ExpenseShare
	if request.Body.Code != nil {
		res, err := s.queries.GetExpenseShareByCode(ctx, data.GetExpenseShareByCodeParams{
			Code: *request.Body.Code,
		})
		if err != nil {
			return nil, err
		}
		if time.Since(res.ExpenseShareCode.Created.Time) > config.Values.ExpenseShareCodeTtl.Duration {
			return nil, fmt.Errorf("Expense share code expired")
		}
		expense_share = res.ExpenseShare
	} else if request.Body.DefaultName != nil {
		expense_share, err = s.queries.CreateExpenseShare(ctx, data.CreateExpenseShareParams{
			DefaultName: *request.Body.DefaultName,
		})
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("Join code or default name required")
	}
	if request.Body.DisplayName == "" {
		return nil, fmt.Errorf("`displayName` is required in request body")
	}
	budget_expense_share, err := s.queries.JoinExpenseShare(ctx, data.JoinExpenseShareParams{
		ExpenseShareID: expense_share.ID,
		LoginID:        loginID,
		BudgetID:       int64(budgetID),
		DisplayName:    request.Body.DisplayName,
	})
	if err != nil {
		return nil, err
	}
	// Send the response
	return PostBudgetBudgetIdExpenseShare200JSONResponse{
		Id:          int(expense_share.ID),
		DefaultName: expense_share.DefaultName,
		Name:        budget_expense_share.Name,
		DisplayName: budget_expense_share.DisplayName,
	}, nil
}

func (s ApiServer) GetBudgetBudgetIdExpenseShareExpenseShareIdCode(ctx context.Context, request GetBudgetBudgetIdExpenseShareExpenseShareIdCodeRequestObject) (GetBudgetBudgetIdExpenseShareExpenseShareIdCodeResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetString := request.BudgetId
	budgetID, err := strconv.Atoi(budgetString)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	expenseShareString := request.ExpenseShareId
	expenseShareID, err := strconv.Atoi(expenseShareString)
	if err != nil {
		return nil, fmt.Errorf("Invalid expenseShare id: %w", err)
	}
	// Query the database
	expenseShareCode, err := s.queries.CreateExpenseShareCode(ctx, data.CreateExpenseShareCodeParams{
		BudgetID: int64(budgetID),
		LoginID: int64(loginID),
		ExpenseShareID: int64(expenseShareID),
		Code: uuid.NewString(),
	})
	// Send the response
	return GetBudgetBudgetIdExpenseShareExpenseShareIdCode200JSONResponse{
		Code: expenseShareCode.Code,
	}, nil
}

func (s ApiServer) PutBudgetBudgetIdExpenseShareExpenseShareId(ctx context.Context, request PutBudgetBudgetIdExpenseShareExpenseShareIdRequestObject) (PutBudgetBudgetIdExpenseShareExpenseShareIdResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetString := request.BudgetId
	budgetID, err := strconv.Atoi(budgetString)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	expenseShareString := request.ExpenseShareId
	expenseShareID, err := strconv.Atoi(expenseShareString)
	if err != nil {
		return nil, fmt.Errorf("Invalid expenseShare id: %w", err)
	}
	if request.Body == nil || request.Body.Name == "" {
		return nil, fmt.Errorf("`name` is required in request body")
	}
	if request.Body.DisplayName == "" {
		return nil, fmt.Errorf("`displayName` is required in request body")
	}
	// Query the database
	budgetExpenseShare, err := s.queries.UpdateExpenseShare(ctx, data.UpdateExpenseShareParams{
		Name:           request.Body.Name,
		DisplayName:    request.Body.DisplayName,
		ExpenseShareID: int64(expenseShareID),
		BudgetID:       int64(budgetID),
		LoginID:        loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send the response
	return PutBudgetBudgetIdExpenseShareExpenseShareId200JSONResponse{
		Id:          int(budgetExpenseShare.ExpenseShareID),
		Name:        budgetExpenseShare.Name,
		DisplayName: budgetExpenseShare.DisplayName,
	}, nil
}

func (s ApiServer) DeleteBudgetBudgetIdExpenseShareExpenseShareId(ctx context.Context, request DeleteBudgetBudgetIdExpenseShareExpenseShareIdRequestObject) (DeleteBudgetBudgetIdExpenseShareExpenseShareIdResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetString := request.BudgetId
	budgetID, err := strconv.Atoi(budgetString)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	expenseShareString := request.ExpenseShareId
	expenseShareID, err := strconv.Atoi(expenseShareString)
	if err != nil {
		return nil, fmt.Errorf("Invalid expenseShare id: %w", err)
	}
	// Query the database
	err = s.queries.LeaveExpenseShare(ctx, data.LeaveExpenseShareParams{
		ExpenseShareID: int64(expenseShareID),
		BudgetID: int64(budgetID),
		LoginID: int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	// Send the response
	return DeleteBudgetBudgetIdExpenseShareExpenseShareId204Response{}, nil
}

func (s ApiServer) PostBudgetBudgetIdExpenseShareExpenseShareIdTrx(ctx context.Context, request PostBudgetBudgetIdExpenseShareExpenseShareIdTrxRequestObject) (PostBudgetBudgetIdExpenseShareExpenseShareIdTrxResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetString := request.BudgetId
	budgetID, err := strconv.Atoi(budgetString)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	expenseShareString := request.ExpenseShareId
	expenseShareID, err := strconv.Atoi(expenseShareString)
	if err != nil {
		return nil, fmt.Errorf("Invalid expenseShare id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`trxId` is required in request body")
	}
	// Query the db
	row, err := s.queries.PublishExpenseShareTrx(ctx, data.PublishExpenseShareTrxParams{
		ExpenseShareID: int64(expenseShareID),
		TrxID:          int64(request.Body.TrxId),
		BudgetID:       int64(budgetID),
		LoginID:        loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send the response
	return PostBudgetBudgetIdExpenseShareExpenseShareIdTrx201JSONResponse{
		Id:               int(row.ID),
		ExpenseShareId:   int(row.ExpenseShareID),
		TrxId:            int(row.TrxID.Int64),
		PayeeName:        row.PayeeName,
		Date:             row.Date.Format(time.RFC3339),
		TotalOutflow:     int(row.TotalOutflow),
		TotalInflow:      int(row.TotalInflow),
		RequestedOutflow: int(row.RequestedOutflow),
		RequestedInflow:  int(row.RequestedInflow),
		Note:             row.Note,
	}, nil
}

func (s ApiServer) GetBudgetBudgetIdExpenseShareExpenseShareId(ctx context.Context, request GetBudgetBudgetIdExpenseShareExpenseShareIdRequestObject) (GetBudgetBudgetIdExpenseShareExpenseShareIdResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetString := request.BudgetId
	budgetID, err := strconv.Atoi(budgetString)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	expenseShareString := request.ExpenseShareId
	expenseShareID, err := strconv.Atoi(expenseShareString)
	if err != nil {
		return nil, fmt.Errorf("Invalid expenseShare id: %w", err)
	}
	// Member ids double as the membership proof: a member always sees at
	// least themselves, so an empty list means not your budget / not a member.
	members, err := s.queries.ListExpenseShareMembers(ctx, data.ListExpenseShareMembersParams{
		ExpenseShareID: int64(expenseShareID),
		BudgetID:       int64(budgetID),
		LoginID:        loginID,
	})
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, fmt.Errorf("Expense share not found for this budget")
	}
	share, err := s.queries.GetBudgetExpenseShare(ctx, data.GetBudgetExpenseShareParams{
		ExpenseShareID: int64(expenseShareID),
		BudgetID:       int64(budgetID),
		LoginID:        loginID,
	})
	if err != nil {
		return nil, err
	}
	// Flat trx x splits x own-lines rows, gated on the same membership.
	rows, err := s.queries.ListExpenseShareTrxDetails(ctx, data.ListExpenseShareTrxDetailsParams{
		ExpenseShareID: int64(expenseShareID),
		BudgetID:       sql.NullInt64{Int64: int64(budgetID), Valid: true},
		LoginID:        loginID,
	})
	if err != nil {
		return nil, err
	}
	transactions := groupExpenseShareTrx(int64(budgetID), members, rows)
	// Per-member net balances come straight from SQL.
	balances, err := s.queries.ListExpenseShareBalances(ctx, data.ListExpenseShareBalancesParams{
		ExpenseShareID: int64(expenseShareID),
		BudgetID:       int64(budgetID),
		LoginID:        loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send the response
	return GetBudgetBudgetIdExpenseShareExpenseShareId200JSONResponse{
		Summary:      summarizeExpenseShareTrx(share, members, transactions, balances),
		Transactions: transactions,
	}, nil
}

// summarizeExpenseShareTrx builds the share summary: identity and counts
// plus the SQL-computed per-member net balances (positive = others owe this
// member; negative = this member owes).
func summarizeExpenseShareTrx(share data.GetBudgetExpenseShareRow, members []data.ListExpenseShareMembersRow, transactions []ExpenseShareTrxDetail, balances []data.ListExpenseShareBalancesRow) ExpenseShareSummary {
	summary := ExpenseShareSummary{
		Id:               int(share.ExpenseShareID),
		Name:             share.Name,
		DisplayName:      share.DisplayName,
		MemberCount:      len(members),
		TransactionCount: len(transactions),
		Balances:         make([]ExpenseShareBalance, 0, len(balances)),
	}
	for _, b := range balances {
		summary.Balances = append(summary.Balances, ExpenseShareBalance{
			BudgetId:    int(b.BudgetID),
			DisplayName: b.DisplayName,
			Balance:     int(b.Balance),
		})
	}
	return summary
}

// groupExpenseShareTrx nests flat detail rows into trx -> splits -> own lines,
// synthesizing default splits (requested / owing members, leftover cents dealt
// in budget-id order) for members without a stored split.
func groupExpenseShareTrx(budgetID int64, members []data.ListExpenseShareMembersRow, rows []data.ListExpenseShareTrxDetailsRow) []ExpenseShareTrxDetail {
	resp := make([]ExpenseShareTrxDetail, 0)
	byTrx := make(map[int64]int)
	memberIDs := make([]int64, 0, len(members))
	displayNames := make(map[int64]string, len(members))
	for _, m := range members {
		memberIDs = append(memberIDs, m.BudgetID)
		displayNames[m.BudgetID] = m.DisplayName
	}
	isMember := make(map[int64]bool, len(members))
	for _, m := range memberIDs {
		isMember[m] = true
	}
	for _, row := range rows {
		i, ok := byTrx[row.EstID]
		if !ok {
			i = len(resp)
			byTrx[row.EstID] = i
			resp = append(resp, ExpenseShareTrxDetail{
				Id:                   int(row.EstID),
				ExpenseShareId:         int(row.ExpenseShareID),
				TrxId:                  int(row.SourceTrxID.Int64),
				AccountId:              nullInt64ToInt(row.SourceAccountID),
				PublisherBudgetId:      nullInt64ToInt(row.PublisherBudgetID),
				PublisherDisplayName:   nullStringToPtr(row.PublisherDisplayName),
				PayeeName:              row.PayeeName,
				Date:                   row.Date.Format(time.RFC3339),
				TotalOutflow:           int(row.TotalOutflow),
				TotalInflow:            int(row.TotalInflow),
				RequestedOutflow:       int(row.RequestedOutflow),
				RequestedInflow:        int(row.RequestedInflow),
				Note:                   row.Note,
				Splits:                 make([]ExpenseShareTrxSplit, 0),
			})
		}
		detail := &resp[i]
		// Collect stored splits (current members only; splits orphaned by a
		// deleted budget or left behind by an ex-member are ignored) and own
		// lines; rows fan out as trx x splits x own-lines, so dedupe by id.
		var split *ExpenseShareTrxSplit
		if row.SplitID.Valid && row.SplitBudgetID.Valid && isMember[row.SplitBudgetID.Int64] {
			split = findExpenseShareSplit(detail.Splits, row.SplitBudgetID.Int64)
			if split == nil {
				detail.Splits = append(detail.Splits, ExpenseShareTrxSplit{
					BudgetId:     int(row.SplitBudgetID.Int64),
					DisplayName:  displayNames[row.SplitBudgetID.Int64],
					SplitOutflow: int(row.SplitOutflow.Int64),
					SplitInflow:  int(row.SplitInflow.Int64),
				})
				split = &detail.Splits[len(detail.Splits)-1]
			}
		}
		if row.LineID.Valid && row.SplitBudgetID.Int64 == budgetID && split != nil {
			if split.Lines == nil {
				split.Lines = &[]ExpenseShareTrxSplitLine{}
			}
			if !hasExpenseShareLine(*split.Lines, row.LineID.Int64) {
				*split.Lines = append(*split.Lines, ExpenseShareTrxSplitLine{
					Id:              int(row.LineID.Int64),
					CategoryId:      nullInt64ToInt(row.CategoryID),
					CategoryName:    nullStringToPtr(row.CategoryName),
					DestAccountId:   nullInt64ToInt(row.DestAccountID),
					DestAccountName: nullStringToPtr(row.DestAccountName),
					Outflow:         int(row.LineOutflow.Int64),
					Inflow:          int(row.LineInflow.Int64),
				})
			}
		}
	}
	// Fill defaults for members without a stored split. Members arrive
	// ordered by budget id, which also fixes leftover-cent distribution order.
	// Non-publishers share the requested amount; the publisher keeps the
	// unshared remainder (total - requested).
	for i := range resp {
		detail := &resp[i]
		var publisher int64
		publisherKnown := false
		if detail.PublisherBudgetId != nil {
			publisher, publisherKnown = int64(*detail.PublisherBudgetId), true
		}
		owing := make([]int64, 0, len(memberIDs))
		for _, m := range memberIDs {
			if findExpenseShareSplit(detail.Splits, m) == nil {
				owing = append(owing, m)
			}
		}
		// Own split gets an empty (non-nil) lines list when uncategorized so
		// the frontend can distinguish "mine, uncategorized" from "other's".
		markOwnSplit := func() {
			if s := findExpenseShareSplit(detail.Splits, budgetID); s != nil && s.Lines == nil {
				s.Lines = &[]ExpenseShareTrxSplitLine{}
			}
		}
		if len(owing) == 0 {
			markOwnSplit()
			continue
		}
		outflow := detail.TotalInflow == 0
		requested := int64(detail.RequestedOutflow)
		if !outflow {
			requested = int64(detail.RequestedInflow)
		}
		sharing := make([]int64, 0, len(owing))
		for _, m := range owing {
			if !publisherKnown || m != publisher {
				sharing = append(sharing, m)
			}
		}
		if len(sharing) > 0 {
			base := requested / int64(len(sharing))
			extra := requested % int64(len(sharing))
			for j, m := range sharing {
				amount := base
				if int64(j) < extra {
					amount++
				}
				split := ExpenseShareTrxSplit{BudgetId: int(m), DisplayName: displayNames[m], IsDefault: true}
				if outflow {
					split.SplitOutflow = int(amount)
				} else {
					split.SplitInflow = int(amount)
				}
				detail.Splits = append(detail.Splits, split)
			}
		}
		// The publisher keeps the unshared remainder (total - requested,
		// possibly 0). Only listed while they are still a share member;
		// departed or deleted publishers vanish from splits entirely.
		if publisherKnown && isMember[publisher] && findExpenseShareSplit(detail.Splits, publisher) == nil {
			kept := int64(detail.TotalOutflow) - int64(detail.RequestedOutflow)
			if !outflow {
				kept = int64(detail.TotalInflow) - int64(detail.RequestedInflow)
			}
			split := ExpenseShareTrxSplit{BudgetId: int(publisher), DisplayName: displayNames[publisher], IsDefault: true}
			if outflow {
				split.SplitOutflow = int(kept)
			} else {
				split.SplitInflow = int(kept)
			}
			detail.Splits = append(detail.Splits, split)
		}
		markOwnSplit()
	}
	return resp
}

func (s ApiServer) PutBudgetBudgetIdExpenseShareExpenseShareIdTrxTrxIdSplits(ctx context.Context, request PutBudgetBudgetIdExpenseShareExpenseShareIdTrxTrxIdSplitsRequestObject) (PutBudgetBudgetIdExpenseShareExpenseShareIdTrxTrxIdSplitsResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetID, err := strconv.Atoi(request.BudgetId)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	expenseShareID, err := strconv.Atoi(request.ExpenseShareId)
	if err != nil {
		return nil, fmt.Errorf("Invalid expenseShare id: %w", err)
	}
	trxID, err := strconv.Atoi(request.TrxId)
	if err != nil {
		return nil, fmt.Errorf("Invalid trx id: %w", err)
	}
	if request.Body == nil || len(request.Body.Splits) == 0 {
		return nil, fmt.Errorf("`splits` is required in request body")
	}
	// Any member may edit splits; an empty member list means not your
	// budget / not a member.
	members, err := s.queries.ListExpenseShareMembers(ctx, data.ListExpenseShareMembersParams{
		ExpenseShareID: int64(expenseShareID),
		BudgetID:       int64(budgetID),
		LoginID:        loginID,
	})
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, fmt.Errorf("Expense share not found for this budget")
	}
	trx, err := s.queries.GetExpenseShareTrxById(ctx, data.GetExpenseShareTrxByIdParams{
		ExpenseShareID: int64(expenseShareID),
		TrxID:          int64(trxID),
		BudgetID:       int64(budgetID),
		LoginID:        loginID,
	})
	if err != nil {
		return nil, err
	}
	outflow := trx.TotalInflow == 0
	var requested, total int64
	if outflow {
		requested, total = trx.RequestedOutflow, trx.TotalOutflow
	} else {
		requested, total = trx.RequestedInflow, trx.TotalInflow
	}
	// The publisher keeps the unshared remainder (total - requested) and its
	// split is managed automatically, so it is never accepted from the client.
	// A departed or deleted publisher has no split at all.
	var publisher int64
	publisherKnown := false
	if trx.BudgetID.Valid {
		publisher, publisherKnown = trx.BudgetID.Int64, true
	}
	// Validate every split before touching the database.
	isMember := make(map[int]bool, len(members))
	for _, m := range members {
		isMember[int(m.BudgetID)] = true
	}
	if publisherKnown && !isMember[int(publisher)] {
		publisherKnown = false
	}
	seen := make(map[int]bool, len(request.Body.Splits))
	var sum int64
	for _, split := range request.Body.Splits {
		if publisherKnown && int64(split.BudgetId) == publisher {
			return nil, fmt.Errorf("Split for the publishing budget %d is set automatically", split.BudgetId)
		}
		if !isMember[split.BudgetId] {
			return nil, fmt.Errorf("Budget %d is not a member of this expense share", split.BudgetId)
		}
		if seen[split.BudgetId] {
			return nil, fmt.Errorf("Duplicate split for budget %d", split.BudgetId)
		}
		seen[split.BudgetId] = true
		if split.Outflow < 0 || split.Inflow < 0 {
			return nil, fmt.Errorf("Split amounts must not be negative")
		}
		if outflow && split.Inflow != 0 {
			return nil, fmt.Errorf("Inflow splits are not allowed on an outflow transaction")
		}
		if !outflow && split.Outflow != 0 {
			return nil, fmt.Errorf("Outflow splits are not allowed on an inflow transaction")
		}
		// A zero split is allowed for members who are not paying.
		sum += int64(split.Outflow + split.Inflow)
	}
	// Every non-publisher member needs a stored split; otherwise the read path
	// would synthesize a default on top of the stored ones.
	for _, m := range members {
		if publisherKnown && m.BudgetID == publisher {
			continue
		}
		if !seen[int(m.BudgetID)] {
			return nil, fmt.Errorf("Missing split for budget %d", m.BudgetID)
		}
	}
	if sum != requested {
		return nil, fmt.Errorf("Splits total %d != requested %d", sum, requested)
	}
	// Replace all stored splits atomically.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	txQueries := s.queries.WithTx(tx)
	if err := txQueries.DeleteExpenseShareSplits(ctx, data.DeleteExpenseShareSplitsParams{
		TrxID: int64(trxID),
	}); err != nil {
		return nil, err
	}
	for _, split := range request.Body.Splits {
		if _, err := txQueries.CreateExpenseShareSplit(ctx, data.CreateExpenseShareSplitParams{
			TrxID:          int64(trxID),
			ExpenseShareID: int64(expenseShareID),
			BudgetID:       sql.NullInt64{Int64: int64(split.BudgetId), Valid: true},
			SplitOutflow:   int64(split.Outflow),
			SplitInflow:    int64(split.Inflow),
		}); err != nil {
			return nil, err
		}
	}
	if publisherKnown && total-requested > 0 {
		var splitOutflow, splitInflow int64
		if outflow {
			splitOutflow = total - requested
		} else {
			splitInflow = total - requested
		}
		if _, err := txQueries.CreateExpenseShareSplit(ctx, data.CreateExpenseShareSplitParams{
			TrxID:          int64(trxID),
			ExpenseShareID: int64(expenseShareID),
			BudgetID:       sql.NullInt64{Int64: publisher, Valid: true},
			SplitOutflow:   splitOutflow,
			SplitInflow:    splitInflow,
		}); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	// Re-read and return the updated transaction detail.
	rows, err := s.queries.ListExpenseShareTrxDetails(ctx, data.ListExpenseShareTrxDetailsParams{
		ExpenseShareID: int64(expenseShareID),
		BudgetID:       sql.NullInt64{Int64: int64(budgetID), Valid: true},
		LoginID:        loginID,
	})
	if err != nil {
		return nil, err
	}
	for _, detail := range groupExpenseShareTrx(int64(budgetID), members, rows) {
		if detail.Id == trxID {
			return PutBudgetBudgetIdExpenseShareExpenseShareIdTrxTrxIdSplits200JSONResponse(detail), nil
		}
	}
	return nil, fmt.Errorf("Expense share transaction not found")
}

func (s ApiServer) PutBudgetBudgetIdExpenseShareExpenseShareIdTrxTrxIdLines(ctx context.Context, request PutBudgetBudgetIdExpenseShareExpenseShareIdTrxTrxIdLinesRequestObject) (PutBudgetBudgetIdExpenseShareExpenseShareIdTrxTrxIdLinesResponseObject, error) {
	// Collect params
	loginID, err := getLoginID(ctx)
	if err != nil {
		return nil, err
	}
	budgetID, err := strconv.Atoi(request.BudgetId)
	if err != nil {
		return nil, fmt.Errorf("Invalid budget id: %w", err)
	}
	expenseShareID, err := strconv.Atoi(request.ExpenseShareId)
	if err != nil {
		return nil, fmt.Errorf("Invalid expenseShare id: %w", err)
	}
	trxID, err := strconv.Atoi(request.TrxId)
	if err != nil {
		return nil, fmt.Errorf("Invalid trx id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`lines` is required in request body")
	}
	// Any member may categorize their own split; an empty member list means
	// not your budget / not a member.
	members, err := s.queries.ListExpenseShareMembers(ctx, data.ListExpenseShareMembersParams{
		ExpenseShareID: int64(expenseShareID),
		BudgetID:       int64(budgetID),
		LoginID:        loginID,
	})
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, fmt.Errorf("Expense share not found for this budget")
	}
	trx, err := s.queries.GetExpenseShareTrxById(ctx, data.GetExpenseShareTrxByIdParams{
		ExpenseShareID: int64(expenseShareID),
		TrxID:          int64(trxID),
		BudgetID:       int64(budgetID),
		LoginID:        loginID,
	})
	if err != nil {
		return nil, err
	}
	// The publisher categorizes through the source transaction's own lines, so
	// split lines are only for the other members (recording both would double
	// count in the publisher's budget views).
	if trx.BudgetID.Valid && trx.BudgetID.Int64 == int64(budgetID) {
		return nil, fmt.Errorf("The publishing budget categorizes through the source transaction")
	}
	outflow := trx.TotalInflow == 0
	// Targets must live in this budget; fetch the valid ids once.
	categories, err := s.queries.ListCategories(ctx, data.ListCategoriesParams{
		LoginID:  loginID,
		BudgetID: int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	validCategory := make(map[int]bool, len(categories))
	for _, c := range categories {
		validCategory[int(c.ID)] = true
	}
	accounts, err := s.queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{
		LoginID:  loginID,
		BudgetID: int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	validAccount := make(map[int]bool, len(accounts))
	for _, a := range accounts {
		validAccount[int(a.ID)] = true
	}
	// Validate every line before touching the database.
	for _, line := range request.Body.Lines {
		if (line.CategoryId == nil) == (line.DestAccountId == nil) {
			return nil, fmt.Errorf("Each line needs exactly one of categoryId or destAccountId")
		}
		if line.CategoryId != nil && !validCategory[*line.CategoryId] {
			return nil, fmt.Errorf("Category %d not found in this budget", *line.CategoryId)
		}
		if line.DestAccountId != nil && !validAccount[*line.DestAccountId] {
			return nil, fmt.Errorf("Account %d not found in this budget", *line.DestAccountId)
		}
		if line.Outflow < 0 || line.Inflow < 0 {
			return nil, fmt.Errorf("Line amounts must not be negative")
		}
		if outflow && (line.Inflow != 0 || line.Outflow == 0) {
			return nil, fmt.Errorf("Lines on an outflow transaction need an outflow amount")
		}
		if !outflow && (line.Outflow != 0 || line.Inflow == 0) {
			return nil, fmt.Errorf("Lines on an inflow transaction need an inflow amount")
		}
	}
	// Replace this budget's lines atomically, materializing a stored split
	// from the (possibly defaulted) amounts first when there isn't one yet.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	txQueries := s.queries.WithTx(tx)
	split, err := txQueries.GetExpenseShareSplit(ctx, data.GetExpenseShareSplitParams{
		TrxID:    int64(trxID),
		BudgetID: sql.NullInt64{Int64: int64(budgetID), Valid: true},
	})
	if err == sql.ErrNoRows {
		// Read through the tx: the open transaction holds this test's only
		// migrated :memory: connection, so a pooled query could land on a
		// fresh empty connection.
		rows, err := txQueries.ListExpenseShareTrxDetails(ctx, data.ListExpenseShareTrxDetailsParams{
			ExpenseShareID: int64(expenseShareID),
			BudgetID:       sql.NullInt64{Int64: int64(budgetID), Valid: true},
			LoginID:        loginID,
		})
		if err != nil {
			return nil, err
		}
		var own *ExpenseShareTrxSplit
		for _, detail := range groupExpenseShareTrx(int64(budgetID), members, rows) {
			if detail.Id == trxID {
				own = findExpenseShareSplit(detail.Splits, int64(budgetID))
				break
			}
		}
		if own == nil {
			return nil, fmt.Errorf("No split found for this budget")
		}
		split, err = txQueries.CreateExpenseShareSplit(ctx, data.CreateExpenseShareSplitParams{
			TrxID:          int64(trxID),
			ExpenseShareID: int64(expenseShareID),
			BudgetID:       sql.NullInt64{Int64: int64(budgetID), Valid: true},
			SplitOutflow:   int64(own.SplitOutflow),
			SplitInflow:    int64(own.SplitInflow),
		})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	// A zero split has nothing to categorize.
	splitAmount := split.SplitOutflow
	if !outflow {
		splitAmount = split.SplitInflow
	}
	if splitAmount == 0 && len(request.Body.Lines) > 0 {
		return nil, fmt.Errorf("Cannot categorize a zero split")
	}
	if err := txQueries.DeleteExpenseShareSplitLines(ctx, data.DeleteExpenseShareSplitLinesParams{
		SplitID: split.ID,
	}); err != nil {
		return nil, err
	}
	for _, line := range request.Body.Lines {
		if _, err := txQueries.CreateExpenseShareSplitLine(ctx, data.CreateExpenseShareSplitLineParams{
			BudgetID:      int64(budgetID),
			SplitID:       split.ID,
			DestAccountID: intToNullInt64(line.DestAccountId),
			CategoryID:    intToNullInt64(line.CategoryId),
			Outflow:       int64(line.Outflow),
			Inflow:        int64(line.Inflow),
		}); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	// Re-read and return the updated transaction detail.
	rows, err := s.queries.ListExpenseShareTrxDetails(ctx, data.ListExpenseShareTrxDetailsParams{
		ExpenseShareID: int64(expenseShareID),
		BudgetID:       sql.NullInt64{Int64: int64(budgetID), Valid: true},
		LoginID:        loginID,
	})
	if err != nil {
		return nil, err
	}
	for _, detail := range groupExpenseShareTrx(int64(budgetID), members, rows) {
		if detail.Id == trxID {
			return PutBudgetBudgetIdExpenseShareExpenseShareIdTrxTrxIdLines200JSONResponse(detail), nil
		}
	}
	return nil, fmt.Errorf("Expense share transaction not found")
}

func findExpenseShareSplit(splits []ExpenseShareTrxSplit, budgetID int64) *ExpenseShareTrxSplit {
	for i := range splits {
		if splits[i].BudgetId == int(budgetID) {
			return &splits[i]
		}
	}
	return nil
}

func hasExpenseShareLine(lines []ExpenseShareTrxSplitLine, id int64) bool {
	for _, l := range lines {
		if l.Id == int(id) {
			return true
		}
	}
	return false
}
