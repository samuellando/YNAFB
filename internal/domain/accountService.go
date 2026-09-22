package domain

type AccountService struct {
	repo AccountRepository
	trxService *TrxService
}

func NewAccountService(repo AccountRepository, trxService *TrxService) *AccountService {
	return &AccountService{repo: repo, trxService: trxService}
}

func (s *AccountService) SetTrxService(trxService *TrxService) {
	s.trxService = trxService
}
