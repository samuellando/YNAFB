package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"samuellando.com/YNAFB/internal/auth"
	"samuellando.com/YNAFB/internal/data"
	dbutil "samuellando.com/YNAFB/internal/db"
	"samuellando.com/YNAFB/internal/domain"
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
	domain  *domain.DomainService
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
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() { 
		db.Close() 
		cancel()
	})
	queries := data.New(db)
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
	service := domain.NewDomainService(queries)
	h := middleware.Authenticator(api.Handler(api.NewStrictHandler(server, nil)))
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", h))
	return &testServer{mux: mux, domain: service, ctx: ctx, loginID: login.ID, token: token}
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

func relMonth(m int) time.Time {
	return time.Now().AddDate(0, m, 0)
}

func sRelMonth(m int) string {
	return time.Now().AddDate(0, m, 0).Format("2006-01")
}

