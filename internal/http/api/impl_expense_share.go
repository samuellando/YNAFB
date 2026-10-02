package api

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"samuellando.com/YNAFB/internal/domain"
)

// marshalMembership translates an ExpenseShareMembership to the wire shape.
// IDs resolve via object navigation (Budget stays the sole root); the
// conversion lives here in the API layer, never as domain FK getters.
func marshalMembership(share *domain.ExpenseShareMembership) ExpenseShareMembership {
	return ExpenseShareMembership{
		Id:             share.ID(),
		ExpenseShareId: share.Share().ID(),
		BudgetId:       share.Budget().ID(),
		Name:           share.Name(),
		DisplayName:    share.DisplayName(),
	}
}

// marshalMembers translates active share members to the wire shape.
func marshalMembers(members []*domain.ShareMember) []ShareMember {
	marshalled := make([]ShareMember, 0, len(members))
	for _, m := range members {
		marshalled = append(marshalled, ShareMember{
			BudgetId:    m.Budget.ID(),
			DisplayName: m.DisplayName,
		})
	}
	return marshalled
}

// marshalSummary translates a caller-relative balance and pairwise nets to
// the wire shape.
func marshalSummary(balance int, balances []domain.MemberBalance) ExpenseShareSummary {
	marshalled := make([]ExpenseShareMemberBalance, 0, len(balances))
	for _, b := range balances {
		marshalled = append(marshalled, ExpenseShareMemberBalance{
			BudgetId:    b.Budget.ID(),
			DisplayName: b.DisplayName,
			Balance:     b.Balance,
		})
	}
	return ExpenseShareSummary{
		Balance:        balance,
		MemberBalances: marshalled,
	}
}

// marshalCode translates an ExpenseShareCode to the wire shape.
func marshalCode(code *domain.ExpenseShareCode) ExpenseShareCode {
	return ExpenseShareCode{
		Code:    code.Code(),
		Expires: code.Expires().Format(time.RFC3339),
	}
}

// categorizationCategory translates a SplitCategorization's category object
// to the nullable wire ID/name. The entity exposes objects (not scalar FKs),
// so the conversion lives here in the API layer.
func categorizationCategory(cat *domain.SplitCategorization) (categoryID *int, categoryName *string) {
	if category, err := cat.Category(); err == nil {
		id := category.ID()
		name := category.Name()
		categoryID = &id
		categoryName = &name
	}
	return categoryID, categoryName
}

// marshalCategorization translates a SplitCategorization to the wire shape.
func marshalCategorization(cat *domain.SplitCategorization) SplitCategorization {
	categoryID, categoryName := categorizationCategory(cat)
	return SplitCategorization{
		Id:           cat.ID(),
		CategoryId:   categoryID,
		CategoryName: categoryName,
		Outflow:      cat.Outflow(),
		Inflow:       cat.Inflow(),
	}
}

// marshalShareTransaction translates an ExpenseShareTransaction view to the
// wire shape. Split lines, settlement and categorizations resolve live;
// absent relations stay omitted until the domain layer fills them in.
func marshalShareTransaction(trx *domain.ExpenseShareTransaction, categorizations []SplitCategorization) ExpenseShareTransaction {
	resp := ExpenseShareTransaction{
		SourceBudgetId:          trx.SourceBudget().ID(),
		SourceBudgetDisplayName: trx.SourceBudgetDisplayName(),
		TrxId:                   trx.SourceTrx().ID(),
		PayeeName:               trx.PayeeName(),
		Date:                    trx.Date(),
		TotalOutflow:            trx.TotalOutflow(),
		TotalInflow:             trx.TotalInflow(),
		Note:                    trx.Note(),
	}
	if categorizations != nil {
		resp.MyCategorizations = &categorizations
	}
	return resp
}

// resolveCategorizationCategory loads the category for a required wire ID.
func resolveCategorizationCategory(ctx context.Context, budget *domain.Budget, categoryID int) (*domain.Category, error) {
	return budget.GetCategory(ctx, categoryID)
}

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
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	shares, err := budget.ListExpenseShares(ctx)
	if err != nil {
		return nil, err
	}
	// Send response
	resp := GetBudgetBudgetIdExpenseShare200JSONResponse{}
	for _, share := range shares {
		resp = append(resp, marshalMembership(share))
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
	if request.Body == nil {
		return nil, fmt.Errorf("`name` and `displayName` are required in request body")
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	share, err := budget.CreateExpenseShare(ctx, request.Body.Name, request.Body.DisplayName, request.Body.DefaultName)
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdExpenseShare200JSONResponse(marshalMembership(share)), nil
}

func (s ApiServer) PostBudgetBudgetIdExpenseShareJoin(ctx context.Context, request PostBudgetBudgetIdExpenseShareJoinRequestObject) (PostBudgetBudgetIdExpenseShareJoinResponseObject, error) {
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
	if request.Body == nil {
		return nil, fmt.Errorf("`code`, `name` and `displayName` are required in request body")
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	share, err := budget.JoinExpenseShare(ctx, request.Body.Code, request.Body.Name, request.Body.DisplayName)
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdExpenseShareJoin200JSONResponse(marshalMembership(share)), nil
}

func (s ApiServer) GetBudgetBudgetIdExpenseShareShareId(ctx context.Context, request GetBudgetBudgetIdExpenseShareShareIdRequestObject) (GetBudgetBudgetIdExpenseShareShareIdResponseObject, error) {
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
	shareString := request.ShareId
	shareID, err := strconv.Atoi(shareString)
	if err != nil {
		return nil, fmt.Errorf("Invalid share id: %w", err)
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	share, err := budget.GetExpenseShare(ctx, shareID)
	if err != nil {
		return nil, err
	}
	transactions, err := share.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	members, err := share.Members(ctx)
	if err != nil {
		return nil, err
	}
	balance, balances, err := share.Balance(ctx)
	if err != nil {
		return nil, err
	}
	// Send response
	trxs := make([]ExpenseShareTransaction, 0, len(transactions))
	for _, trx := range transactions {
		cats, err := trx.MyCategorizations(ctx)
		if err != nil {
			return nil, err
		}
		marshalled := make([]SplitCategorization, 0, len(cats))
		for _, cat := range cats {
			marshalled = append(marshalled, marshalCategorization(cat))
		}
		trxs = append(trxs, marshalShareTransaction(trx, marshalled))
	}
	return GetBudgetBudgetIdExpenseShareShareId200JSONResponse{
		Membership:   marshalMembership(share),
		Summary:      marshalSummary(balance, balances),
		Members:      marshalMembers(members),
		Transactions: trxs,
	}, nil
}

func (s ApiServer) PutBudgetBudgetIdExpenseShareShareId(ctx context.Context, request PutBudgetBudgetIdExpenseShareShareIdRequestObject) (PutBudgetBudgetIdExpenseShareShareIdResponseObject, error) {
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
	shareString := request.ShareId
	shareID, err := strconv.Atoi(shareString)
	if err != nil {
		return nil, fmt.Errorf("Invalid share id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`name` or `displayName` is required in request body")
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	share, err := budget.GetExpenseShare(ctx, shareID)
	if err != nil {
		return nil, err
	}
	err = share.Update(ctx, request.Body.Name, request.Body.DisplayName)
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdExpenseShareShareId200JSONResponse(marshalMembership(share)), nil
}

func (s ApiServer) DeleteBudgetBudgetIdExpenseShareShareId(ctx context.Context, request DeleteBudgetBudgetIdExpenseShareShareIdRequestObject) (DeleteBudgetBudgetIdExpenseShareShareIdResponseObject, error) {
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
	shareString := request.ShareId
	shareID, err := strconv.Atoi(shareString)
	if err != nil {
		return nil, fmt.Errorf("Invalid share id: %w", err)
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	share, err := budget.GetExpenseShare(ctx, shareID)
	if err != nil {
		return nil, err
	}
	err = share.Leave(ctx)
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdExpenseShareShareId200Response{}, nil
}

func (s ApiServer) PostBudgetBudgetIdExpenseShareShareIdCode(ctx context.Context, request PostBudgetBudgetIdExpenseShareShareIdCodeRequestObject) (PostBudgetBudgetIdExpenseShareShareIdCodeResponseObject, error) {
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
	shareString := request.ShareId
	shareID, err := strconv.Atoi(shareString)
	if err != nil {
		return nil, fmt.Errorf("Invalid share id: %w", err)
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	share, err := budget.GetExpenseShare(ctx, shareID)
	if err != nil {
		return nil, err
	}
	code, err := share.MintCode(ctx)
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdExpenseShareShareIdCode200JSONResponse(marshalCode(code)), nil
}

func (s ApiServer) GetBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLine(ctx context.Context, request GetBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLineRequestObject) (GetBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLineResponseObject, error) {
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
	shareString := request.ShareId
	shareID, err := strconv.Atoi(shareString)
	if err != nil {
		return nil, fmt.Errorf("Invalid share id: %w", err)
	}
	trxString := request.TrxId
	trxID, err := strconv.Atoi(trxString)
	if err != nil {
		return nil, fmt.Errorf("Invalid transaction id: %w", err)
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	share, err := budget.GetExpenseShare(ctx, shareID)
	if err != nil {
		return nil, err
	}
	trx, err := share.GetTransaction(ctx, trxID)
	if err != nil {
		return nil, err
	}
	cats, err := trx.MyCategorizations(ctx)
	if err != nil {
		return nil, err
	}
	// Send response
	resp := GetBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLine200JSONResponse{}
	for _, cat := range cats {
		resp = append(resp, marshalCategorization(cat))
	}
	return resp, nil
}

func (s ApiServer) PostBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLine(ctx context.Context, request PostBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLineRequestObject) (PostBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLineResponseObject, error) {
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
	shareString := request.ShareId
	shareID, err := strconv.Atoi(shareString)
	if err != nil {
		return nil, fmt.Errorf("Invalid share id: %w", err)
	}
	trxString := request.TrxId
	trxID, err := strconv.Atoi(trxString)
	if err != nil {
		return nil, fmt.Errorf("Invalid transaction id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`categoryId`, `outflow` and `inflow` are required in request body")
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	share, err := budget.GetExpenseShare(ctx, shareID)
	if err != nil {
		return nil, err
	}
	trx, err := share.GetTransaction(ctx, trxID)
	if err != nil {
		return nil, err
	}
	category, err := resolveCategorizationCategory(ctx, budget, request.Body.CategoryId)
	if err != nil {
		return nil, err
	}
	cat, err := trx.AddCategorization(ctx, category, request.Body.Outflow, request.Body.Inflow)
	if err != nil {
		return nil, err
	}
	// Send response
	return PostBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLine200JSONResponse(marshalCategorization(cat)), nil
}

func (s ApiServer) DeleteBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLineLineId(ctx context.Context, request DeleteBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLineLineIdRequestObject) (DeleteBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLineLineIdResponseObject, error) {
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
	shareString := request.ShareId
	shareID, err := strconv.Atoi(shareString)
	if err != nil {
		return nil, fmt.Errorf("Invalid share id: %w", err)
	}
	trxString := request.TrxId
	trxID, err := strconv.Atoi(trxString)
	if err != nil {
		return nil, fmt.Errorf("Invalid transaction id: %w", err)
	}
	lineString := request.LineId
	lineID, err := strconv.Atoi(lineString)
	if err != nil {
		return nil, fmt.Errorf("Invalid categorization id: %w", err)
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	share, err := budget.GetExpenseShare(ctx, shareID)
	if err != nil {
		return nil, err
	}
	trx, err := share.GetTransaction(ctx, trxID)
	if err != nil {
		return nil, err
	}
	cat, err := trx.GetCategorization(ctx, lineID)
	if err != nil {
		return nil, err
	}
	err = cat.Delete(ctx)
	if err != nil {
		return nil, err
	}
	return DeleteBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLineLineId200Response{}, nil
}

func (s ApiServer) PutBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLineLineId(ctx context.Context, request PutBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLineLineIdRequestObject) (PutBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLineLineIdResponseObject, error) {
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
	shareString := request.ShareId
	shareID, err := strconv.Atoi(shareString)
	if err != nil {
		return nil, fmt.Errorf("Invalid share id: %w", err)
	}
	trxString := request.TrxId
	trxID, err := strconv.Atoi(trxString)
	if err != nil {
		return nil, fmt.Errorf("Invalid transaction id: %w", err)
	}
	lineString := request.LineId
	lineID, err := strconv.Atoi(lineString)
	if err != nil {
		return nil, fmt.Errorf("Invalid categorization id: %w", err)
	}
	if request.Body == nil {
		return nil, fmt.Errorf("`categoryId`, `outflow` and `inflow` are required in request body")
	}
	// Query the database
	budget, err := s.service.GetBudget(ctx, int(loginID), budgetID)
	if err != nil {
		return nil, err
	}
	share, err := budget.GetExpenseShare(ctx, shareID)
	if err != nil {
		return nil, err
	}
	trx, err := share.GetTransaction(ctx, trxID)
	if err != nil {
		return nil, err
	}
	cat, err := trx.GetCategorization(ctx, lineID)
	if err != nil {
		return nil, err
	}
	category, err := resolveCategorizationCategory(ctx, budget, request.Body.CategoryId)
	if err != nil {
		return nil, err
	}
	err = cat.Update(ctx, category, request.Body.Outflow, request.Body.Inflow)
	if err != nil {
		return nil, err
	}
	// Send response
	return PutBudgetBudgetIdExpenseShareShareIdTransactionTrxIdLineLineId200JSONResponse(marshalCategorization(cat)), nil
}
