package api

import (
	"database/sql"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/db/querycount"
	"samuellando.com/YNAFB/internal/domain"
)

type ApiServer struct {
	service *domain.DomainService
	queries *data.Queries
	db      *sql.DB
}

var _ StrictServerInterface = (*ApiServer)(nil)

func NewServer(db *sql.DB) ApiServer {
	qdb := querycount.New(db)
	queries := data.New(qdb)
	service := domain.NewDomainService(queries)
	return ApiServer{
		service: service,
		queries: data.New(qdb),
		db:      db,
	}
}
