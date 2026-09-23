# YNAFB — application architecture

YNAFB ("You Need A F** Budget") is a single-binary personal budgeting app:
a Go HTTP server on `:8080` backed by SQLite, serving both the JSON API and
the embedded React SPA. One `login` owns many `budget`s; every query is
scoped by `(login_id, budget_id)`.

```
┌─ frontend/ (React + TS SPA, Vite) ──────────────┐
│ openapi-fetch client over generated schema.d.ts │  dev: :5173, proxies /api + /auth → :8080
└──────────────────────┬──────────────────────────┘
                       │  HttpOnly jwt cookie, credentials: "include"
┌─ cmd/ynafb-server (:8080) ─────────────────────┐
│  /auth/* ── handler (login.go: signup/in/out)  │
│  /api/v1/* ── middleware.Authenticator          │
│              → StrictHandler → ApiServer        │
│  /* ── spaHandler (embedded web/ bundle)        │
└──────────────────────┬──────────────────────────┘
┌─ internal/domain/ ─────────────────────────────┐
│  Rich object graph rooted at Budget (see        │
│  docs/domain.md). Single DomainService +        │
│  DomainRepository. Request-scoped identity map  │
│  + result cache (internal/cache/).              │
└──────────────────────┬──────────────────────────┘
┌─ internal/data/ (sqlc) ─┴──────────────────────────┐
│  *data.Queries generated from queries/*.sql +   │
│  migrations/*.sql. Raw rows only, no logic.     │
└──────────────────────┬──────────────────────────┘
                  SQLite (ynafb.db, goose via dbutil.Open)
```

## API definition (`api.yml`, repo root)

`api.yml` (OpenAPI 3.1.0) is the single source of truth for `/budget/...`
endpoints and the `components.schemas` models. The `/auth` endpoints are
hand-written and not in the spec.

- **Backend**: `internal/http/api/server.gen.go` from oapi-codegen
  (`api.go:3` generate directive, config in `cfg.yml`): models,
  `ServerInterface` routing on a `ServeMux`, and the
  `StrictServerInterface` the hand-written `impl_*.go` files implement via
  `ApiServer{service, queries, db}` (`server.go`). Generated Go identifiers
  derive from method + path (`GET /budget/{budgetId}` → `GetBudgetBudgetId`),
  so path renames ripple through generated types.
- **Frontend**: `frontend/src/lib/api/schema.d.ts` from openapi-typescript
  (`npm run gen:api`); `client.ts` wraps openapi-fetch (`baseUrl: "/api/v1"`,
  credentials included).
- Change flow: edit `api.yml` → `go generate ./...` (backend) and/or
  `npm run gen:api` (frontend) → update hand-written consumers.

## Request flow (API)

`Authenticator` validates the `jwt` cookie, refreshes it, and injects
`loginID` into the request context (`getLoginID`). Each `impl_*.go` handler
then: parse ids → `service.GetBudget(ctx, loginID, budgetID)` → walk the
domain graph (`budget.GetAccount` → `account.GetTransaction` →
`trx.GetLine`, …) → return a typed response object. Statement import
additionally wraps the work in a SQL tx via
`service.WithRepo(queries.WithTx(tx))` so multi-row imports stay atomic.

## Persistence (`internal/data/` + `queries/` + `migrations/`)

- `queries/*.sql` holds one named query per operation with
  `@login_id`/`@budget_id` scoping baked in; `sqlc.yml` generates
  `internal/data/` (`internal/data/*.sql.go` is gitignored). List/get queries that feed object graphs
  `JOIN` related tables (category/group names, line details) so the domain
  hydrates in one round trip — see `queries/goal.sql`, `trx.sql`,
  `category.sql`, `allocation.sql`.
- `migrations/*.sql` is the schema (incl. `CHECK` constraints the domain
  relies on, e.g. goal `type`/`end_date`, trx inflow-xor-outflow).
- `internal/db/types` maps `UNIX_EPOCH_INTEGER` ↔ `UnixTime`/`NullUnixTime`;
  money is integer cents on the wire.

## Cross-cutting pieces

- **Cache** (`internal/cache/`): per-request-context instance map
  (`cache.Get` identity map) + result map (`cache.Result`, cleared by
  `InvalidateResults`). Contexts are request-scoped, so entries never leak
  across requests (or logins).
- **Importer** (`internal/importer/`): statement parsers per bank/format
  (`desjardins/`, `wealthsimple/…`, registered in `parser.go`) producing a
  `statement.Statement`, consumed by `Account.ImportStatement`
  (get-or-create payee per entry, one trx each, no lines).
- **Month math** lives in the domain (`budgetCalculate.go`,
  `month.ts` on the frontend): `YYYY-MM` strings, UTC, pure year/month
  arithmetic — no `Date` day math.

## Frontend (`frontend/`)

Thin pages (`src/pages/`) composing feature components
(`src/components/budget/`), shared primitives (`src/components/ui/`),
React Query keyed `['budget-month', id, month]` etc., view helpers in
`src/lib/` (`budgetView.ts`, `month.ts`). Conventions that matter:
display-until-click inline edits, `DialogShell` for dialogs, listbox
semantics for pickers, deliberate query-invalidation scope. Full detail
lives in `AGENTS.md`.

## Build, run, verify

- `make build` — frontend bundle (`build.outDir` →
  `cmd/ynafb-server/web/`, embedded via `//go:embed all:web`) + Go binary.
  `web/` is gitignored; `make run`'s `ensure-web` rebuilds it if missing.
- `make run` — API server `:8080` + Vite `:5173` (proxies `/api`, `/auth`).
- Backend: `go build ./...`, `go test -count=1 ./internal/...`,
  `go test -count=1 ./internal/data/`.
- Frontend: `npm run build` (`tsc -b` + vite), `npm run lint` (oxlint).
