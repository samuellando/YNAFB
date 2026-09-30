# YNAFB Budgeting

Envelope budgeting: income funds category allocations per month; spending draws from Available.

## Language

**Budget**:
The root scope for all budgeting; owned by one login.
_Avoid_: book, workspace

**Account**:
A real-world money holder (checking, savings) with a balance from transaction totals.
_Avoid_: wallet

**Payee**:
The counterparty on every transaction; required even if unknown.
_Avoid_: merchant, payer

**Transaction**:
A dated total (inflow XOR outflow) with a payee in one account; lines may partially categorize it.
_Avoid_: entry

**Spend line**:
A transaction line assigning outflow/inflow to a Category.
_Avoid_: categorization

**Transfer line**:
A transaction line moving money to a destination Account; budget-neutral.
_Avoid_: move

**Income line**:
A transaction line marking inflow as budgetable income.
_Avoid_: inflow, revenue

**Uncategorized**:
The global gap between transaction totals and line totals.
_Avoid_: unassigned, leftover

**Category**:
A spending purpose; optionally in a Group, never named `income`.
_Avoid_: tag, label

**Category Group**:
An optional grouping for categories; delete orphans categories to Other.
_Avoid_: group, folder

**Allocation**:
Money assigned to a Category for one month; absent implies zero, may be negative.
_Avoid_: budget, funding

**Available**:
Allocated minus spent plus prior positive carry; what a category can still spend.
_Avoid_: balance, remaining

**ReadyToAssign**:
Income to date minus spent-before, allocations, and positive carry; what is left to allocate.
_Avoid_: free money, to-budget

**Goal**:
At most one target per Category: monthly, save-by-date, or refill-to-amount.
_Avoid_: target, objective

**Payee Default Line**:
A stored split suggestion (percent) for a payee; currently unchecked and unapplied.
_Avoid_: template, rule

**Reconciliation**:
Idempotent per-account checkpoint that owned transactions and incoming transfer lines existed as of a date.
_Avoid_: lock, close

**Mirror transaction**:
A read-only swapped-total view of a transfer in its destination account; never stored.
_Avoid_: copy, double-entry
