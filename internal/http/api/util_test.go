package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/auth"
	dbutil "samuellando.com/YNAFB/internal/db"
	"samuellando.com/YNAFB/internal/db/types"
	"samuellando.com/YNAFB/internal/http/api"
	"samuellando.com/YNAFB/internal/http/middleware"
	"samuellando.com/YNAFB/internal/importer"
	"samuellando.com/YNAFB/internal/importer/testparser"

	"github.com/pressly/goose/v3"
)

func init() {
	importer.Register(testparser.Parser{})
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), v); err != nil {
		t.Fatalf("decode body %q: %v", w.Body.String(), err)
	}
}

type testServer struct {
	mux     *http.ServeMux
	queries *data.Queries
	db      *sql.DB
	ctx     context.Context
	loginID int64
	token   string
}

func setupTestServer(t *testing.T) *testServer {
	t.Helper()
	goose.SetLogger(goose.NopLogger())
	db, err := dbutil.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	queries := data.New(db)
	ctx := context.Background()
	login, err := queries.CreateLogin(ctx, data.CreateLoginParams{Username: "tester", Password: "password"})
	if err != nil {
		t.Fatal(err)
	}
	token, err := auth.GenerateJWT(strconv.Itoa(int(login.ID)))
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	server := api.NewServer(db)
	h := middleware.Authenticator(api.Handler(api.NewStrictHandler(server, nil)))
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", h))
	return &testServer{mux: mux, queries: queries, db: db, ctx: ctx, loginID: login.ID, token: token}
}

func (ts *testServer) doReq(t *testing.T, method, path string, body any, want int) *httptest.ResponseRecorder {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		rd = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.AddCookie(&http.Cookie{Name: "jwt", Value: ts.token})
	w := httptest.NewRecorder()
	ts.mux.ServeHTTP(w, req)
	if w.Code != want {
		t.Fatalf("%s %s: status = %d, want %d, body = %q", method, path, w.Code, want, w.Body.String())
	}
	return w
}

func (ts *testServer) newBudget(t *testing.T, name string) data.Budget {
	t.Helper()
	b, err := ts.queries.CreateBudget(ts.ctx, data.CreateBudgetParams{LoginID: ts.loginID, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func (ts *testServer) newAccount(t *testing.T, budget data.Budget, name string) data.Account {
	t.Helper()
	a, err := ts.queries.CreateAccount(ts.ctx, data.CreateAccountParams{BudgetID: budget.ID, LoginID: budget.LoginID, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return data.Account{ID: a.ID, BudgetID: a.BudgetID, Name: a.Name}
}

func (ts *testServer) newPayee(t *testing.T, budget data.Budget, name string) data.Payee {
	t.Helper()
	p, err := ts.queries.CreatePayee(ts.ctx, data.CreatePayeeParams{BudgetID: budget.ID, LoginID: budget.LoginID, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func (ts *testServer) newCategory(t *testing.T, budget data.Budget, name string) data.Category {
	t.Helper()
	c, err := ts.queries.CreateCategory(ts.ctx, data.CreateCategoryParams{BudgetID: budget.ID, LoginID: budget.LoginID, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func (ts *testServer) newCategoryGroup(t *testing.T, budget data.Budget, name string) int64 {
	t.Helper()
	id, err := ts.queries.CreateCategoryGroup(ts.ctx, data.CreateCategoryGroupParams{BudgetID: budget.ID, LoginID: budget.LoginID, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func (ts *testServer) newGoal(t *testing.T, budget data.Budget, category data.Category) data.Goal {
	t.Helper()
	g, err := ts.queries.CreateGoal(ts.ctx, data.CreateGoalParams{
		BudgetID:   budget.ID,
		LoginID:    budget.LoginID,
		Type:       "monthly",
		StartDate:  types.UnixTime{Time: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
		EndDate:    types.NullUnixTime{},
		CategoryID: category.ID,
		Amount:     5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func (ts *testServer) newDefaultLine(t *testing.T, budget data.Budget, payee data.Payee) data.PayeeDefaultLine {
	t.Helper()
	l, err := ts.queries.CreatePayeeDefaultLine(ts.ctx, data.CreatePayeeDefaultLineParams{
		BudgetID: budget.ID,
		LoginID:  budget.LoginID,
		PayeeID:  payee.ID,
		Income:   true,
		Percent:  100,
	})
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func (ts *testServer) newTrx(t *testing.T, budget data.Budget, account data.Account, payee data.Payee) data.Trx {
	t.Helper()
	trx, err := ts.queries.CreateTrx(ts.ctx, data.CreateTrxParams{
		BudgetID:     budget.ID,
		LoginID:      budget.LoginID,
		AccountID:    account.ID,
		PayeeID:      payee.ID,
		Date:         types.UnixTime{Time: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)},
		TotalOutflow: 1000,
		TotalInflow:  0,
		Note:         "",
	})
	if err != nil {
		t.Fatal(err)
	}
	return trx
}

func (ts *testServer) newTrxLine(t *testing.T, budget data.Budget, trx data.Trx) data.TrxLine {
	t.Helper()
	l, err := ts.queries.CreateTrxLine(ts.ctx, data.CreateTrxLineParams{
		BudgetID: budget.ID,
		LoginID:  budget.LoginID,
		TrxID:    trx.ID,
		Income:   true,
		Outflow:  0,
		Inflow:   500,
	})
	if err != nil {
		t.Fatal(err)
	}
	return l
}