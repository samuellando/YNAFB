package budget

import (
	"context"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/domain/allocation"
	"samuellando.com/YNAFB/internal/domain/trx"
)

type Budget struct {
	repo Repository
	trxService *trx.Service
	allocationService *allocation.Service
	row  data.Budget
}

func (b *Budget) ID() int {
	return int(b.row.ID)
}

func (b *Budget) Name() string {
	return b.row.Name
}

type Service struct {
	repo       Repository
	trxService *trx.Service
	allocationService *allocation.Service
}

func NewService(repo Repository, trxService *trx.Service, allocationService *allocation.Service) *Service {
	return &Service{repo: repo, trxService: trxService, allocationService: allocationService}
}

func (s *Service) List(ctx context.Context, loginID int) ([]*Budget, error) {
	rows, err := s.repo.ListBudgets(ctx, data.ListBudgetsParams{
		LoginID: int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	budgets := make([]*Budget, len(rows))
	for i, row := range rows {
		budgets[i] = &Budget{
			repo: s.repo,
			trxService: s.trxService,
			row:  row,
		}
	}
	return budgets, nil
}

func (s *Service) Create(ctx context.Context, loginID int, name string) (*Budget, error) {
	row, err := s.repo.CreateBudget(ctx, data.CreateBudgetParams{
		LoginID: int64(loginID),
		Name:    name,
	})
	if err != nil {
		return nil, err
	}
	return &Budget{
		repo: s.repo,
			trxService: s.trxService,
			allocationService: s.allocationService,
		row:  row,
	}, nil
}

func (s *Service) Get(ctx context.Context, loginID, budgetID int) (*Budget, error) {
	row, err := s.repo.GetBudget(ctx, data.GetBudgetParams{
		ID:      int64(budgetID),
		LoginID: int64(loginID),
	})
	if err != nil {
		return nil, err
	}
	return &Budget{
		repo: s.repo,
			trxService: s.trxService,
			allocationService: s.allocationService,
		row:  row,
	}, nil
}

func (b *Budget) Update(ctx context.Context, name string) error {
	row, err := b.repo.UpdateBudget(ctx, data.UpdateBudgetParams{
		LoginID: b.row.LoginID,
		ID:      b.row.ID,
		Name:    name,
	})
	b.row = row
	return err
}

func (b *Budget) Delete(ctx context.Context) error {
	return b.repo.DeleteBudget(ctx, data.DeleteBudgetParams{
		LoginID: b.row.LoginID,
		ID:      b.row.ID,
	})
}
