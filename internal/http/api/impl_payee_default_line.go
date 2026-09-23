package api

import (
	"context"
	"fmt"
	"strconv"

	"samuellando.com/YNAFB/internal/domain"
)

// defaultLineRefs unwraps a line's related objects, tolerating absence.
func defaultLineRefs(line *domain.PayeeDefaultLine) (dest *domain.Account, category *domain.Category) {
	if d, err := line.DestAccount(); err == nil {
		dest = d
	}
	if c, err := line.Category(); err == nil {
		category = c
	}
	return dest, category
}

// resolveDefaultLineRefs loads the related objects for optional wire IDs.
func resolveDefaultLineRefs(ctx context.Context, budget *domain.Budget, destID, categoryID *int) (*domain.Account, *domain.Category, error) {
	var dest *domain.Account
	if destID != nil {
		var err error
		dest, err = budget.GetAccount(ctx, *destID)
		if err != nil {
			return nil, nil, err
		}
	}
	var category *domain.Category
	if categoryID != nil {
		var err error
		category, err = budget.GetCategory(ctx, *categoryID)
		if err != nil {
			return nil, nil, err
		}
	}
	return dest, category, nil
}

// marshalDefaultLine translates a PayeeDefaultLine and its related objects
// to the wire shape. The entity exposes objects (not scalar FKs), so the
// conversion lives here in the API layer.
func marshalDefaultLine(line *domain.PayeeDefaultLine, dest *domain.Account, category *domain.Category) PayeeDefaultLine {
	resp := PayeeDefaultLine{
		Id:      line.ID(),
		PayeeId: line.Payee().ID(),
		Income:  line.Income(),
		Percent: line.Percent(),
	}
	if dest != nil {
		id := dest.ID()
		name := dest.Name()
		resp.DestAccountId = &id
		resp.DestAccountName = &name
	}
	if category != nil {
		id := category.ID()
		name := category.Name()
		resp.CategoryId = &id
		resp.CategoryName = &name
	}
	return resp
}

func (s ApiServer) GetBudgetBudgetIdPayeePayeeIdDefaultLine(ctx context.Context, request GetBudgetBudgetIdPayeePayeeIdDefaultLineRequestObject) (GetBudgetBudgetIdPayeePayeeIdDefaultLineResponseObject, error) {
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
	payeeString := request.PayeeId
	payeeID, err := strconv.Atoi(payeeString)
	if err != nil {
		return nil, fmt.Errorf("Invalid payee id: %w", err)
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	payee, err := budget.GetPayee(ctx, payeeID)
	if err != nil {
		return nil, err
	}
	lines, err := payee.DefaultLines(ctx)
	if err != nil {
		return nil, err
	}
	// Send response
	resp := GetBudgetBudgetIdPayeePayeeIdDefaultLine200JSONResponse{}
	for _, line := range lines {
		dest, category := defaultLineRefs(line)
		resp = append(resp, marshalDefaultLine(line, dest, category))
	}
	return resp, nil
}

func (s ApiServer) PostBudgetBudgetIdPayeePayeeIdDefaultLine(ctx context.Context, request PostBudgetBudgetIdPayeePayeeIdDefaultLineRequestObject) (PostBudgetBudgetIdPayeePayeeIdDefaultLineResponseObject, error) {
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
	payeeString := request.PayeeId
	payeeID, err := strconv.Atoi(payeeString)
	if err != nil {
		return nil, fmt.Errorf("Invalid payee id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`percent` is required in request body")
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	payee, err := budget.GetPayee(ctx, payeeID)
	if err != nil {
		return nil, err
	}
	dest, category, err := resolveDefaultLineRefs(ctx, budget, request.Body.DestAccountId, request.Body.CategoryId)
	if err != nil {
		return nil, err
	}
	line, err := payee.AddDefaultLine(ctx, dest, category, BoolPtrToBool(request.Body.Income), request.Body.Percent)
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdPayeePayeeIdDefaultLine200JSONResponse(marshalDefaultLine(line, dest, category)), nil
}

func (s ApiServer) PutBudgetBudgetIdPayeePayeeIdDefaultLineId(ctx context.Context, request PutBudgetBudgetIdPayeePayeeIdDefaultLineIdRequestObject) (PutBudgetBudgetIdPayeePayeeIdDefaultLineIdResponseObject, error) {
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
	payeeString := request.PayeeId
	payeeID, err := strconv.Atoi(payeeString)
	if err != nil {
		return nil, fmt.Errorf("Invalid payee id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid default line id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`percent` is required in request body")
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	payee, err := budget.GetPayee(ctx, payeeID)
	if err != nil {
		return nil, err
	}
	line, err := payee.GetDefaultLine(ctx, id)
	if err != nil {
		return nil, err
	}
	err = line.Update(ctx, request.Body.DestAccountId, request.Body.CategoryId, BoolPtrToBool(request.Body.Income), request.Body.Percent)
	if err != nil {
		return nil, err
	}
	// The line's related objects are stale after update, so resolve fresh
	// objects for the response.
	dest, category, err := resolveDefaultLineRefs(ctx, budget, request.Body.DestAccountId, request.Body.CategoryId)
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdPayeePayeeIdDefaultLineId200JSONResponse(marshalDefaultLine(line, dest, category)), nil
}

func (s ApiServer) DeleteBudgetBudgetIdPayeePayeeIdDefaultLineId(ctx context.Context, request DeleteBudgetBudgetIdPayeePayeeIdDefaultLineIdRequestObject) (DeleteBudgetBudgetIdPayeePayeeIdDefaultLineIdResponseObject, error) {
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
	payeeString := request.PayeeId
	payeeID, err := strconv.Atoi(payeeString)
	if err != nil {
		return nil, fmt.Errorf("Invalid payee id: %w", err)
	}
	idString := request.Id
	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, fmt.Errorf("Invalid default line id: %w", err)
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	payee, err := budget.GetPayee(ctx, payeeID)
	if err != nil {
		return nil, err
	}
	line, err := payee.GetDefaultLine(ctx, id)
	if err != nil {
		return nil, err
	}
	err = line.Delete(ctx)
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdPayeePayeeIdDefaultLineId200Response{}, nil
}
