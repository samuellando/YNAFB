# Shared ("mirror") accounts — implementation plan

Status: **draft for review** — nothing below is implemented yet.

## 1. Overview

YNAFB is a single-user CLI, but a household can contain two budgets (two users) in
the same database. When one person pays for a shared expense, the money is really
split between the two budgets. Today there is no mechanism to reflect that split in
the *other* budget.

This feature adds a **shared settlement account** ("mirror account") that links two
budgets so a shared expense paid by one person automatically shows up on the other
person's side, where they categorize their own portion — and the account's balance
tells you who owes whom.

It deliberately builds on the **existing transfer primitive** (`transaction_category`
with `other_account` set). No new transaction types are introduced; the change is
mostly about *ownership* of category lines, *scope* of account views, and the sign
convention applied to the shared account.

### The core idea

- An account **M** is created in one budget ("owner") and **shared** into a second
  budget.
- M has no native transactions. It only accumulates **transfer lines**
  (`other_account = M`) from both users' real accounts (chequing, etc.).
- Every category line gains an **owner budget**. A transfer into M owned by budget B
  means "the other person owes B money" from B's point of view.
- M's balance is the net settlement and is **exact opposites** in the two budgets:
  budget B sees `+X` for transfers *it* made and `−X` for transfers the *other*
  budget made.

### Example

1. Alice pays $40 for groceries out of her chequing.
2. She splits the transaction: `$20 → Groceries` (her category) and
   `$20 → transfer to M` (the shared account).
3. Bob sees, in his view of M, a single `$20 from Alice` entry.
4. Bob categorizes that $20 into whatever category he wants (e.g. his own Groceries),
   attributed to **his** budget.
5. Bob does the same for shared expenses he pays; Alice sees those on her end.
6. M's balance shows the running settlement. If Alice's M reads `+$20`, she is owed
   $20; Bob's M reads `−$20`.

## 2. Feature behaviour (user-facing)

- `account share create <account_id> <budget_id>` — share an account into another
  budget. Raw integer IDs (no name resolution), consistent with existing CLI
  conventions for IDs.
- `account share delete <account_id>` — remove the share.
- `account share list` — show all shares with account + budget names.
- The shared account appears in the target budget's `account list` (marked as
  shared) with its balance flipped relative to the owner's view.
- `--other-account <name>` resolves shared accounts, so a user in the target budget
  can transfer *into* M from their own real accounts.
- `categorize` on a shared account shows each incoming transfer as a single
  "`$X from <account>`" line; the user assigns their portion with their own
  categories.
- Reconciliation works per budget; a shared account can be reconciled independently
  in each budget.
- Once a transaction is reconciled in an account, it is protected from edits/deletes
  in that account, but the other user may still add their own categorization lines.

## 3. Schema changes (higher-level)

Three tables change, all in one new migration `migrations/00008_account_share.sql`.

### 3.1 New table: `account_share`

| column   | type    | notes                                              |
|----------|---------|----------------------------------------------------|
| `id`     | INTEGER | PK                                                 |
| `account`| INTEGER | FK → account, `ON DELETE CASCADE`, **UNIQUE**      |
| `budget` | INTEGER | FK → budget, `ON DELETE CASCADE`                   |

- `UNIQUE(account)` → an account is shared into **at most one** other budget.
- No self-share: budget ≠ the account's own budget. Enforced in the CLI (SQLite
  `CHECK` cannot reference another table).

### 3.2 Modified: `transaction_category` — add owner `budget`

Rebuilt table (same technique as `00005_income.sql`) adding one column:

- `budget INTEGER NOT NULL REFERENCES budget(id) ON DELETE CASCADE` — the **owner**
  budget of this category/transfer/income line.

Existing rows backfilled to the transaction's account budget (all historical lines
were created in their own budget, so this is exact).

This column is what lets the *other* budget attach its own category line to a
transaction that lives in a different budget.

### 3.3 Modified: `reconciliation` — add `budget` to the key

Rebuilt to `(budget, account, transaction)` PK, backfilled from the account's
budget. Required so a shared account can be reconciled independently in each budget
(Alice's reconciliation of M is distinct from Bob's).

## 4. Query changes

All under `queries/`, followed by `sqlc generate` (the `internal/data/` package is generated;
never hand-edit).

- **`share.sql`** (new): create/delete/list shares, look up the partner budget for an
  account, and `IsTransactionReconciled(budget, account, transaction)`.
- **`transaction_category.sql`**: add owner `budget` to create/update.
- **`account_summary.sql`**: M's balance and reconciled balance use the sign rule
  (see §1); reconciled subqueries keyed by `budget`.
- **`transaction.sql`**: `ListAccountTransactions` becomes owner-scoped and
  shared-aware — for a shared account it returns transfer lines into M (from either
  budget) plus the viewing budget's own category lines, flipped for the sharing
  budget.
- **`reconcile.sql`**: reconcile and as-of balance take a `budget`.
- **`budget.sql`**: income and uncategorized amounts scoped by owner `budget` rather
  than the transaction's account budget, so a user's categorization in a shared
  context lands in their own reports.
- **`list.sql` / `lookup.sql`**: account listing and by-name resolution include
  shared accounts (so `--other-account M` works from the sharing budget).

## 5. Application-layer changes (`cmd/ynafb/main.go`)

- **`account share`** subcommand (create/delete/list).
- **`account list` / `show` / `categorize` / `reconcile`**: resolve and render shared
  accounts; the "viewing budget" drives the sign flip.
- **`transaction category create` / `update`**: write owner `budget` = the acting
  budget.
- **`categorize`**: for a shared account, the queue shows each pulled transfer as a
  single `"$X from <account>"` entry (the transfer amount, not the source
  transaction's total), and the user assigns their portion with their own categories.
  "Totals match" is evaluated per owner against their portion, not the full source
  transaction.
- **Reconciliation protection** (per-account, allows cross-categorization):
  - `transaction update` / `delete` → blocked when reconciled in the transaction's
    own account.
  - `transaction category update` / `delete` → blocked when reconciled in the line's
    owner's account.
  - `transaction category create` → allowed (cross-categorization).
  - `categorize` → skips reconciled transactions.

**No copy hooks.** Because views are computed at query time, creating/importing a
transaction and adding a transfer to M requires no extra writes, no sync, and no
cascade logic.

## 6. Complexities & edge cases

1. **Sign convention.** The heart of the feature. M's balance from budget B is
   `Σ(own transfers) − Σ(other budget's transfers)`. This must be applied in every
   balance/reconciled-balance/listing query that touches M. Getting a sign wrong
   silently corrupts the settlement balance.

2. **Owner attribution for non-category lines.** A category line's owner is normally
   derivable (category → its budget), but **income** and **transfer** lines created
   on a cross-budget transaction are not — hence the explicit owner `budget` column.
   Without it, Bob marking a pulled entry as "income" would be attributed to Alice's
   budget.

3. **Cross-categorization without double counting.** Alice's $40 transaction carries
   her lines (Groceries $20, transfer M $20) *and* Bob's line (his Groceries $20).
   Budget reports must count only lines owned by the reporting budget. The existing
   category-scoped spending queries mostly hold because category → budget, but
   income/uncategorized queries need re-scoping.

4. **Reconciliation keyed by budget.** The `reconciliation` table gains a budget
   dimension. Reconcile queries, as-of balances, and reconciled-balance subqueries
   must all pass budget through, or the two users' reconciliations collide.

5. **Protection matrix.** Reconciled-protection is per-account and intentionally
   asymmetric: edits/deletes are blocked, but adding a *new* category line by the
   other user is allowed. This is the easiest place to over-block (annoying) or
   under-block (allows tampering).

6. **Shared-account name resolution & collisions.** `--other-account M` and
   `account show M` must resolve the shared account in the target budget. If the
   target budget has its own account with the same name, resolution is ambiguous —
   need a deterministic rule (e.g. local account wins, or a distinct namespace for
   shared accounts) and a clear error.

7. **Account deletion / budget deletion.** FK cascades: deleting an account removes
   its share row; deleting a budget removes shares pointing at it and category lines
   it owns. Verify no orphaned share rows or dangling `other_account` references.

8. **`account import`.** Imports write transactions into a real account. If that
   account is shared, imported transfers into M automatically appear on the other
   side via the query-time view — no import-time hook, but the import flow must not
   accidentally resolve M or flip signs.

9. **"Remaining" / totals-match in `categorize`.** For shared accounts each user
   only accounts for their portion, so the strict "remaining must be zero" check must
   be relaxed for shared context while staying strict for normal accounts.

10. **Migration backfill.** Both rebuilt tables need correct backfill of existing
    rows; the `internal/data/` package must be regenerated or the build fails (per AGENTS.md).

## 7. Testing strategy

- Extend `cmd/ynafb/main_test.go` with end-to-end scenarios against a temp SQLite DB:
  - create share → transfer from each side → assert M balances are exact opposites;
  - categorize the pulled entry from the other budget → assert it lands in the other
    budget's reports and not the owner's;
  - reconcile M in each budget independently;
  - reconciled-transaction protection (blocked edit/delete, allowed cross-categorize);
  - literal-spacing assertions in `account show`/`list` updated for shared rows.
- No unit tests under parser subpackages (per AGENTS.md).
- Manual verification of balance math with the example in §1.

## 8. Open questions for review

1. Name/CLI verb: `account share` vs `account mirror`? (Draft uses `share`.)
2. Shared-account naming in the target budget: same name + "shared" marker in
   `account list`, or a distinct namespace? Draft: same name, marked, local-name wins
   on collision.
3. Should `account share create` also validate that the account currently has no
   native transactions (i.e. it's a pure settlement account), or allow sharing any
   account? Draft: allow sharing any account; transfers are the intended use.