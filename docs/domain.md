# Domain package (`internal/domain/`) — structure and rules

The domain package is a rich object graph rooted at `Budget`. The HTTP layer
(`internal/http/api/`) walks the graph; it contains no business logic beyond
translating objects to/from wire shapes. These rules were established during
the service-layer refactor — follow them so the package stays consistent.

## Layout

- One file per entity: `budget.go`, `account.go`, `category.go`,
  `categoryGroup.go`, `allocation.go`, `goal.go`, `payee.go`,
  `payeeDefaultLine.go`, `trx.go`, `trxLine.go`.
- `domainService.go` — the single `DomainService` entry point.
- `domainRepository.go` — the single `DomainRepository` interface (all sqlc
  queries the domain needs; `*data.Queries` satisfies it).
- `budgetCalculate.go` — month summary/category math over fetched domain
  objects (replaces the old SQL views).
- `util.go` — shared private helpers (`nullInt64FromInt`, month helpers).

## The object graph

- `Budget` is the root and holds the only service pointer
  (`service *DomainService`). Everything else reaches the repo through its
  parent chain (e.g. `c.budget.service.repo`, `l.trx.account.budget...`).
- Entities hold `row data.X` plus parent/relation pointers. No entity holds a
  `*Service`, and no method takes `loginID`/`budgetID` — both are derived
  from the parent (`b.LoginID()`, `c.budget.ID()`).
- `Budget` keeps `ID()`, `LoginID()`, `Name()`. `LoginID()` is the
  authentication scope every query needs; it stays.

## Behavior placement

- Collection CRUD hangs off the owner: `Budget.CreatePayee/GetPayee/ListPayees`,
  `Budget.CreateCategory/GetCategory/ListCategories`,
  `Budget.CreateAccount/GetAccount/ListAccounts`,
  `Budget.ListAllocations`, `Budget.ListGoals`, `Budget.ListCategoryGroups…`.
- Child writes hang off their parent: `Trx.AddLine/GetLine`,
  `Account.CreateTransaction/GetTransaction/ListTransactions`,
  `Category.GetGoal`, `Category.SetAllocation`, `Category.Create` (goals),
  `Payee.AddDefaultLine/GetDefaultLine/DefaultLines`.
- Root entry points (they need an explicit `loginID`) are methods on
  `*DomainService`: `ListBudgets`, `CreateBudget`, `GetBudget`
  (constructor: `NewDomainService`). `WithRepo` returns a tx-scoped copy for
  transactional flows (statement import).

## Accessors: objects, not foreign keys

- Entities expose their own `ID()` only. Never add `GroupID()`,
  `CategoryID()`, … callers navigate objects (`category.Group().ID()`).
- Non-nullable relations return a plain pointer: `Goal.Category()`,
  `Allocation.Category()`, `Trx.Account()`, `TrxLine.Trx()`.
- Nullable relations return `(*T, error)`: `TrxLine.Category()`,
  `TrxLine.DestinationAccount()`, `Category.Group()`,
  `PayeeDefaultLine.Category()/DestAccount()`.
- `Allocation` keeps only `category` (no `budget` field) — the parent is
  always present (`category_id NOT NULL`) and carries the budget.

## Construction and the identity map

- Build objects only through `xxxFromRow(ctx, row, parent, …)` helpers,
  which go through `cache.Get(ctx, id, missFunc)`. The same row id yields the
  same pointer — code relies on this (e.g. `Account.ListTransactions`
  filters with `trx.account == a`).
- Partial rows built from join results must still fill `BudgetID` (as the
  payee/category/group partials do). A `BudgetID`-less object cached under an
  id poisons later full loads and mis-scopes `Update`/`Delete`.
- Enrich list/get SQL with `JOIN`s so a list hydrates its relations in one
  round trip (see `ListGoals`, `ListAllocations`, `GetCategory`,
  `ListTrxsAndLines`) instead of N+1 follow-up fetches.
- `trxFromRows` consumes exactly the rows belonging to one transaction:
  count a row only after confirming it matches, and on a cache hit count the
  leading matching rows (`break` on mismatch). Over-consuming silently drops
  transactions from listings (regression test:
  `TestListMultipleTransactionsWithLines`).
- Single-object getters backed by `:many` queries must map “no rows” to
  `sql.ErrNoRows` (see `GetTransaction`) — never return `(nil, nil)`.

## Writes and cache discipline

- Mutations follow the template: `defer cache.InvalidateResults(ctx)`,
  repo call, refresh `row` **and** relation pointers (`Trx.Update` sets
  `t.payee`; `TrxLine.Update`/`Category.Update` set theirs), return.
- `Delete` additionally defers `cache.Delete[*T](ctx, id)`.
- `InvalidateResults` clears only the *result* cache — per-id instance
  objects survive. So collection fields on cached objects must be kept in
  sync by hand: `AddLine` appends to `t.lines`, `TrxLine.Delete` prunes it
  (regression test: `TestTrxLineMutationsVisibleSameContext`). An `Update`
  that leaves a relation pointer stale will serve stale data and stale API
  responses for the rest of the request context.

## HTTP layer contract

- Handlers resolve `service.GetBudget(ctx, loginID, budgetID)` first, then
  walk down (`budget.GetAccount` → `account.GetTransaction` →
  `trx.GetLine`, …). No `*Service` fields on `ApiServer` besides the single
  `service`.
- The entity exposes objects, so nullable wire IDs/names are translated in
  small API-layer helpers next to the handlers (`lineRefIDs`,
  `marshalDefaultLine`, `resolveGroup`/`categoryGroupID`) — never by adding
  getters to the domain.
- Deleting through a path id (account/trx/line) uses that id for scoping;
  entities themselves can no longer move parents (updates never reassign
  `account_id`/`trx_id`).

## Verification

- `go build ./...` must pass; `go test -count=1 ./internal/...` must pass.
  (`go test ./data/` is broken independently — stale month-view tests
  reference view queries that no longer exist since the views moved into
  `budgetCalculate.go`.)
- Keep `gofmt`-clean any file you touch. New domain behavior should come
  with a regression test in `internal/http/api/` (see
  `trx_regression_test.go`).
