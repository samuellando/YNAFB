package domain

type DomainService struct {
	repo DomainRepository
}

func NewDomainService(repo DomainRepository) *DomainService {
	return &DomainService{repo: repo}
}

// WithRepo returns a copy of the service bound to the given repo,
// for use inside a transaction.
func (s *DomainService) WithRepo(repo DomainRepository) *DomainService {
	cp := *s
	cp.repo = repo
	return &cp
}
