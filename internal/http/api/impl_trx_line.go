package api

import (
	"context"
	"fmt"
	"strconv"

	"samuellando.com/YNAFB/internal/domain"
)

// lineRefIDs translates a TrxLine's related objects to the nullable wire IDs.
// The entity exposes objects (not scalar FKs), so the conversion lives here
// in the API layer.
func lineRefIDs(line *domain.TrxLine) (destAccountID, categoryID *int) {
	if dest, err := line.DestinationAccount(); err == nil {
		id := dest.ID()
		destAccountID = &id
	}
	if category, err := line.Category(); err == nil {
		id := category.ID()
		categoryID = &id
	}
	return destAccountID, categoryID
}

func (s ApiServer) PostBudgetBudgetIdAccountAccountIdTransactionTrxIdLine(ctx context.Context, request PostBudgetBudgetIdAccountAccountIdTransactionTrxIdLineRequestObject) (PostBudgetBudgetIdAccountAccountIdTransactionTrxIdLineResponseObject, error) {
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
	accountString := request.AccountId
	accountID, err := strconv.Atoi(accountString)
	if err != nil {
		return nil, fmt.Errorf("Invalid account id: %w", err)
	}
	trxString := request.TrxId
	trxID, err := strconv.Atoi(trxString)
	if err != nil {
		return nil, fmt.Errorf("Invalid transaction id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`outflow` and `inflow` are required in request body")
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	account, err := budget.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	var dest *domain.Account
	if request.Body.DestAccountId != nil {
		dest, err = budget.GetAccount(ctx, *request.Body.DestAccountId)
		if err != nil {
			return nil, err
		}
	}
	var category *domain.Category
	if request.Body.CategoryId != nil {
		category, err = budget.GetCategory(ctx, *request.Body.CategoryId)
		if err != nil {
			return nil, err
		}
	}
	trx, err := account.GetTransaction(ctx, trxID)
	if err != nil {
		return nil, err
	}
	line, err := trx.AddLine(ctx, dest, category, BoolPtrToBool(request.Body.Income), request.Body.Outflow, request.Body.Inflow)
	if err != nil {
		return nil, err
	}
	// Send response
	destAccountID, categoryID := lineRefIDs(line)
	return PostBudgetBudgetIdAccountAccountIdTransactionTrxIdLine200JSONResponse{
		Id:            line.ID(),
		TrxId:         line.Trx().ID(),
		DestAccountId: destAccountID,
		CategoryId:    categoryID,
		Income:        line.IsIncome(),
		Outflow:       line.Outflow(),
		Inflow:        line.Inflow(),
	}, nil
}

func (s ApiServer) PutBudgetBudgetIdAccountAccountIdTransactionTrxIdLineId(ctx context.Context, request PutBudgetBudgetIdAccountAccountIdTransactionTrxIdLineIdRequestObject) (PutBudgetBudgetIdAccountAccountIdTransactionTrxIdLineIdResponseObject, error) {
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
	accountString := request.AccountId
	accountID, err := strconv.Atoi(accountString)
	if err != nil {
		return nil, fmt.Errorf("Invalid account id: %w", err)
	}
	trxString := request.TrxId
	trxID, err := strconv.Atoi(trxString)
	if err != nil {
		return nil, fmt.Errorf("Invalid transaction id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid line id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`outflow` and `inflow` are required in request body")
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	account, err := budget.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	var dest *domain.Account
	if request.Body.DestAccountId != nil {
		dest, err = budget.GetAccount(ctx, *request.Body.DestAccountId)
		if err != nil {
			return nil, err
		}
	}
	var category *domain.Category
	if request.Body.CategoryId != nil {
		category, err = budget.GetCategory(ctx, *request.Body.CategoryId)
		if err != nil {
			return nil, err
		}
	}
	trx, err := account.GetTransaction(ctx, trxID)
	if err != nil {
		return nil, err
	}
	line, err := trx.GetLine(ctx, id)
	if err != nil {
		return nil, err
	}
	err = line.Update(ctx, dest, category, BoolPtrToBool(request.Body.Income), request.Body.Outflow, request.Body.Inflow)
	if err != nil {
		return nil, err
	}
	// Send response
	destAccountID, categoryID := lineRefIDs(line)
	return PutBudgetBudgetIdAccountAccountIdTransactionTrxIdLineId200JSONResponse{
		Id:            line.ID(),
		TrxId:         line.Trx().ID(),
		DestAccountId: destAccountID,
		CategoryId:    categoryID,
		Income:        line.IsIncome(),
		Outflow:       line.Outflow(),
		Inflow:        line.Inflow(),
	}, nil
}

func (s ApiServer) DeleteBudgetBudgetIdAccountAccountIdTransactionTrxIdLineId(ctx context.Context, request DeleteBudgetBudgetIdAccountAccountIdTransactionTrxIdLineIdRequestObject) (DeleteBudgetBudgetIdAccountAccountIdTransactionTrxIdLineIdResponseObject, error) {
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
	accountString := request.AccountId
	accountID, err := strconv.Atoi(accountString)
	if err != nil {
		return nil, fmt.Errorf("Invalid account id: %w", err)
	}
	trxString := request.TrxId
	trxID, err := strconv.Atoi(trxString)
	if err != nil {
		return nil, fmt.Errorf("Invalid transaction id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid line id: %w", err)
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	account, err := budget.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	trx, err := account.GetTransaction(ctx, trxID)
	if err != nil {
		return nil, err
	}
	line, err := trx.GetLine(ctx, id)
	if err != nil {
		return nil, err
	}
	err = line.Delete(ctx)
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdAccountAccountIdTransactionTrxIdLineId200Response{}, nil
}