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
	budget_expense_share, err := s.queries.JoinExpenseShare(ctx, data.JoinExpenseShareParams{
		ExpenseShareID: expense_share.ID,
		LoginID:        loginID,
		BudgetID:       int64(budgetID),
	})
	if err != nil {
		return nil, err
	}
	// Send the response
	return PostBudgetBudgetIdExpenseShare200JSONResponse{
		Id:          int(expense_share.ID),
		DefaultName: expense_share.DefaultName,
		Name:        budget_expense_share.Name,
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
	if request.Body == nil || request.Body.Name == nil || *request.Body.Name == "" {
		return nil, fmt.Errorf("`name` is required in request body")
	}
	// Query the database
	budgetExpenseShare, err := s.queries.UpdateExpenseShare(ctx, data.UpdateExpenseShareParams{
		Name:           *request.Body.Name,
		ExpenseShareID: int64(expenseShareID),
		BudgetID:       int64(budgetID),
		LoginID:        loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send the response
	return PutBudgetBudgetIdExpenseShareExpenseShareId200JSONResponse{
		Id:   int(budgetExpenseShare.ExpenseShareID),
		Name: budgetExpenseShare.Name,
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

func (s ApiServer) GetBudgetBudgetIdExpenseShareExpenseShareIdTrx(ctx context.Context, request GetBudgetBudgetIdExpenseShareExpenseShareIdTrxRequestObject) (GetBudgetBudgetIdExpenseShareExpenseShareIdTrxResponseObject, error) {
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
	// Flat trx x splits x own-lines rows, gated on the same membership.
	rows, err := s.queries.ListExpenseShareTrxDetails(ctx, data.ListExpenseShareTrxDetailsParams{
		ExpenseShareID: int64(expenseShareID),
		BudgetID:       sql.NullInt64{Int64: int64(budgetID), Valid: true},
		LoginID:        loginID,
	})
	if err != nil {
		return nil, err
	}
	// Send the response
	return GetBudgetBudgetIdExpenseShareExpenseShareIdTrx200JSONResponse(
		groupExpenseShareTrx(int64(budgetID), members, rows),
	), nil
}

// groupExpenseShareTrx nests flat detail rows into trx -> splits -> own lines,
// synthesizing default splits (requested / owing members, leftover cents dealt
// in budget-id order) for members without a stored split.
func groupExpenseShareTrx(budgetID int64, members []int64, rows []data.ListExpenseShareTrxDetailsRow) []ExpenseShareTrxDetail {
	resp := make([]ExpenseShareTrxDetail, 0)
	byTrx := make(map[int64]int)
	isMember := make(map[int64]bool, len(members))
	for _, m := range members {
		isMember[m] = true
	}
	for _, row := range rows {
		i, ok := byTrx[row.EstID]
		if !ok {
			i = len(resp)
			byTrx[row.EstID] = i
			resp = append(resp, ExpenseShareTrxDetail{
				Id:               int(row.EstID),
				ExpenseShareId:   int(row.ExpenseShareID),
				TrxId:            int(row.SourceTrxID.Int64),
				PublisherBudgetId: nullInt64ToInt(row.PublisherBudgetID),
				PayeeName:        row.PayeeName,
				Date:             row.Date.Format(time.RFC3339),
				TotalOutflow:     int(row.TotalOutflow),
				TotalInflow:      int(row.TotalInflow),
				RequestedOutflow: int(row.RequestedOutflow),
				RequestedInflow:  int(row.RequestedInflow),
				Note:             row.Note,
				Splits:           make([]ExpenseShareTrxSplit, 0),
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
	// Fill defaults for owing members without a stored split. Members arrive
	// ordered by budget id, which also fixes leftover-cent distribution order.
	for i := range resp {
		detail := &resp[i]
		var publisher int64
		publisherKnown := false
		if detail.PublisherBudgetId != nil {
			publisher, publisherKnown = int64(*detail.PublisherBudgetId), true
		}
		owing := make([]int64, 0, len(members))
		for _, m := range members {
			if publisherKnown && m == publisher {
				continue
			}
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
		requested := int64(detail.RequestedOutflow)
		outflow := true
		if detail.TotalInflow > 0 {
			requested = int64(detail.RequestedInflow)
			outflow = false
		}
		base := requested / int64(len(owing))
		extra := requested % int64(len(owing))
		for j, m := range owing {
			amount := base
			if int64(j) < extra {
				amount++
			}
			split := ExpenseShareTrxSplit{BudgetId: int(m), IsDefault: true}
			if outflow {
				split.SplitOutflow = int(amount)
			} else {
				split.SplitInflow = int(amount)
			}
			detail.Splits = append(detail.Splits, split)
		}
		markOwnSplit()
	}
	return resp
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
