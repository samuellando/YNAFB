package api

import (
	"context"
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
