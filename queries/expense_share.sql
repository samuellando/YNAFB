-- Expense share queries. Ownerless expense_share rows are global; every
-- budget-scoped query proves the caller's membership in the WHERE clause
-- (EXISTS) so scoping lives in SQL, never just in the domain layer.

-- name: GetExpenseShareById :one
SELECT *
FROM expense_share
WHERE id = @id;

-- name: CreateExpenseShare :one
INSERT INTO expense_share (default_name)
VALUES (TRIM(sqlc.arg(default_name)))
RETURNING *;

-- name: DeleteExpenseShare :exec
DELETE FROM expense_share
WHERE id = @id;

-- name: CreateExpenseShareCode :one
INSERT INTO expense_share_code (expense_share_id, code, expires)
VALUES (
  @expense_share_id,
  sqlc.arg(code),
  sqlc.arg(expires)
)
RETURNING *;

-- name: GetExpenseShareByCode :one
SELECT
  sqlc.embed(es),
  c.code AS code,
  c.expires AS expires
FROM expense_share AS es
JOIN expense_share_code AS c ON c.expense_share_id = es.id
WHERE c.code = @code;

-- name: ListMembershipsByBudget :many
SELECT
  sqlc.embed(bes),
  sqlc.embed(es),
  sqlc.embed(b)
FROM budget_expense_share AS bes
JOIN budget AS b ON bes.budget_id = b.id
JOIN expense_share AS es ON es.id = bes.expense_share_id
WHERE b.login_id = @login_id
  AND bes.budget_id = @budget_id
ORDER BY bes.id;

-- name: GetMembershipByShare :one
SELECT
  sqlc.embed(bes),
  sqlc.embed(es),
  sqlc.embed(b)
FROM budget_expense_share AS bes
JOIN budget AS b ON bes.budget_id = b.id
JOIN expense_share AS es ON es.id = bes.expense_share_id
WHERE b.login_id = @login_id
  AND bes.budget_id = @budget_id
  AND bes.expense_share_id = @expense_share_id;

-- name: ListShareMembers :many
SELECT
  sqlc.embed(bes),
  sqlc.embed(es),
  sqlc.embed(b)
FROM budget_expense_share AS bes
JOIN budget AS b ON bes.budget_id = b.id
JOIN expense_share AS es ON es.id = bes.expense_share_id
WHERE bes.expense_share_id = @expense_share_id
  AND EXISTS (
    SELECT 1
    FROM budget_expense_share AS m
    JOIN budget AS cb ON m.budget_id = cb.id
    WHERE m.expense_share_id = @expense_share_id
      AND m.budget_id = @budget_id
      AND cb.login_id = @login_id
  )
ORDER BY bes.id;

-- name: GetShareMember :one
SELECT
  sqlc.embed(bes),
  sqlc.embed(es),
  sqlc.embed(b)
FROM budget_expense_share AS bes
JOIN budget AS b ON bes.budget_id = b.id
JOIN expense_share AS es ON es.id = bes.expense_share_id
WHERE bes.expense_share_id = @expense_share_id
  AND bes.budget_id = @target_budget_id
  AND EXISTS (
    SELECT 1
    FROM budget_expense_share AS m
    JOIN budget AS cb ON m.budget_id = cb.id
    WHERE m.expense_share_id = @expense_share_id
      AND m.budget_id = @budget_id
      AND cb.login_id = @login_id
  );

-- name: CreateMembership :one
INSERT INTO budget_expense_share (name, display_name, expense_share_id, budget_id)
SELECT
  TRIM(sqlc.arg(name)),
  TRIM(sqlc.arg(display_name)),
  @expense_share_id,
  b.id
FROM budget AS b
WHERE b.id = @budget_id
  AND b.login_id = @login_id
RETURNING *;

-- name: UpdateMembership :one
UPDATE budget_expense_share
SET
  name = TRIM(sqlc.arg(name)),
  display_name = TRIM(sqlc.arg(display_name))
WHERE budget_expense_share.expense_share_id = @expense_share_id
  AND budget_expense_share.budget_id = @budget_id
  AND budget_expense_share.budget_id IN (
    SELECT b.id
    FROM budget AS b
    WHERE b.id = @budget_id
      AND b.login_id = @login_id
  )
RETURNING *;

-- name: DeleteMembership :exec
DELETE FROM budget_expense_share
WHERE budget_expense_share.expense_share_id = @expense_share_id
  AND budget_expense_share.budget_id = @budget_id
  AND budget_expense_share.budget_id IN (
    SELECT b.id
    FROM budget AS b
    WHERE b.id = @budget_id
      AND b.login_id = @login_id
  );

-- name: ListShareTransactions :many
SELECT
  sqlc.embed(t),
  a.name AS account_name,
  p.name AS payee_name,
  tl.id AS line_id,
  tl.income AS line_income,
  tl.inflow AS line_inflow,
  tl.outflow AS line_outflow,
  c.id AS category_id,
  c.name AS category_name,
  cg.id AS category_group_id,
  cg.name AS category_group_name,
  da.id AS dest_account_id,
  da.name AS dest_account_name,
  es.default_name AS expense_share_default_name,
  tl.expense_share_id,
  tl.split_budget_expense_share_id,
  tl.dest_budget_expense_share_id,
  sbes.name AS split_budget_expense_share_name,
  sbes.display_name AS split_budget_expense_share_display_name,
  sb.id AS split_budget_id,
  sb.name AS split_budget_name,
  sb.login_id AS split_budget_login_id,
  dbes.name AS dest_budget_expense_share_name,
  dbes.display_name AS dest_budget_expense_share_display_name,
  db.id AS dest_budget_id,
  db.name AS dest_budget_name,
  db.login_id AS dest_budget_login_id,
  CAST(0 AS bool) AS reconciled,
  sm.display_name AS source_member_display_name,
  srcb.id AS source_budget_id,
  srcb.name AS source_budget_name,
  srcb.login_id AS source_budget_login_id
FROM trx AS t
JOIN budget AS srcb ON srcb.id = t.budget_id
JOIN budget_expense_share AS sm ON sm.budget_id = t.budget_id AND sm.expense_share_id = @expense_share_id
JOIN account AS a ON a.id = t.account_id
JOIN payee AS p ON p.id = t.payee_id
LEFT JOIN trx_line AS tl ON tl.trx_id = t.id
LEFT JOIN category AS c ON c.id = tl.category_id
LEFT JOIN category_group AS cg ON cg.id = c.category_group_id
LEFT JOIN account AS da ON da.id = tl.dest_account_id
LEFT JOIN expense_share AS es ON es.id = tl.expense_share_id
LEFT JOIN budget_expense_share AS sbes ON sbes.id = tl.split_budget_expense_share_id
LEFT JOIN budget AS sb ON sb.id = sbes.budget_id
LEFT JOIN budget_expense_share AS dbes ON dbes.id = tl.dest_budget_expense_share_id
LEFT JOIN budget AS db ON db.id = dbes.budget_id
WHERE EXISTS (
    SELECT 1
    FROM trx_line AS sl
    WHERE sl.trx_id = t.id
      AND sl.expense_share_id = @expense_share_id
  )
  AND EXISTS (
    SELECT 1
    FROM budget_expense_share AS m
    JOIN budget AS cb ON m.budget_id = cb.id
    WHERE m.expense_share_id = @expense_share_id
      AND m.budget_id = @budget_id
      AND cb.login_id = @login_id
  )
ORDER BY t.date DESC, t.id DESC;

-- name: GetShareTransaction :many
SELECT
  sqlc.embed(t),
  a.name AS account_name,
  p.name AS payee_name,
  tl.id AS line_id,
  tl.income AS line_income,
  tl.inflow AS line_inflow,
  tl.outflow AS line_outflow,
  c.id AS category_id,
  c.name AS category_name,
  cg.id AS category_group_id,
  cg.name AS category_group_name,
  da.id AS dest_account_id,
  da.name AS dest_account_name,
  es.default_name AS expense_share_default_name,
  tl.expense_share_id,
  tl.split_budget_expense_share_id,
  tl.dest_budget_expense_share_id,
  sbes.name AS split_budget_expense_share_name,
  sbes.display_name AS split_budget_expense_share_display_name,
  sb.id AS split_budget_id,
  sb.name AS split_budget_name,
  sb.login_id AS split_budget_login_id,
  dbes.name AS dest_budget_expense_share_name,
  dbes.display_name AS dest_budget_expense_share_display_name,
  db.id AS dest_budget_id,
  db.name AS dest_budget_name,
  db.login_id AS dest_budget_login_id,
  CAST(0 AS bool) AS reconciled,
  sm.display_name AS source_member_display_name,
  srcb.id AS source_budget_id,
  srcb.name AS source_budget_name,
  srcb.login_id AS source_budget_login_id
FROM trx AS t
JOIN budget AS srcb ON srcb.id = t.budget_id
JOIN budget_expense_share AS sm ON sm.budget_id = t.budget_id AND sm.expense_share_id = @expense_share_id
JOIN account AS a ON a.id = t.account_id
JOIN payee AS p ON p.id = t.payee_id
LEFT JOIN trx_line AS tl ON tl.trx_id = t.id
LEFT JOIN category AS c ON c.id = tl.category_id
LEFT JOIN category_group AS cg ON cg.id = c.category_group_id
LEFT JOIN account AS da ON da.id = tl.dest_account_id
LEFT JOIN expense_share AS es ON es.id = tl.expense_share_id
LEFT JOIN budget_expense_share AS sbes ON sbes.id = tl.split_budget_expense_share_id
LEFT JOIN budget AS sb ON sb.id = sbes.budget_id
LEFT JOIN budget_expense_share AS dbes ON dbes.id = tl.dest_budget_expense_share_id
LEFT JOIN budget AS db ON db.id = dbes.budget_id
WHERE t.id = @trx_id
  AND EXISTS (
    SELECT 1
    FROM trx_line AS sl
    WHERE sl.trx_id = t.id
      AND sl.expense_share_id = @expense_share_id
  )
  AND EXISTS (
    SELECT 1
    FROM budget_expense_share AS m
    JOIN budget AS cb ON m.budget_id = cb.id
    WHERE m.expense_share_id = @expense_share_id
      AND m.budget_id = @budget_id
      AND cb.login_id = @login_id
  )
ORDER BY t.date DESC, t.id DESC;

-- name: ListCategorizations :many
SELECT
  sqlc.embed(s),
  c.name AS category_name,
  cg.id AS category_group_id,
  cg.name AS category_group_name
FROM expense_share_trx_split_line AS s
JOIN budget AS cb ON s.budget_id = cb.id
LEFT JOIN category AS c ON c.id = s.category_id
LEFT JOIN category_group AS cg ON cg.id = c.category_group_id
WHERE cb.login_id = @login_id
  AND s.budget_id = @budget_id
  AND s.expense_share_id = @expense_share_id
  AND s.source_trx_id = @source_trx_id
ORDER BY s.id;

-- name: GetSplitLine :one
SELECT
  sqlc.embed(s),
  c.name AS category_name,
  cg.id AS category_group_id,
  cg.name AS category_group_name
FROM expense_share_trx_split_line AS s
JOIN budget AS cb ON s.budget_id = cb.id
LEFT JOIN category AS c ON c.id = s.category_id
LEFT JOIN category_group AS cg ON cg.id = c.category_group_id
WHERE cb.login_id = @login_id
  AND s.budget_id = @budget_id
  AND s.expense_share_id = @expense_share_id
  AND s.id = @id;

-- name: CreateSplitLine :one
INSERT INTO expense_share_trx_split_line (
  budget_id,
  expense_share_id,
  source_budget_id,
  source_trx_id,
  category_id,
  outflow,
  inflow
)
SELECT
  b.id,
  @expense_share_id,
  @source_budget_id,
  @source_trx_id,
  @category_id,
  @outflow,
  @inflow
FROM budget AS b
WHERE b.id = @budget_id
  AND b.login_id = @login_id
RETURNING *;

-- name: UpdateSplitLine :one
UPDATE expense_share_trx_split_line
SET
  category_id = @category_id,
  outflow = @outflow,
  inflow = @inflow
WHERE expense_share_trx_split_line.id = @id
  AND expense_share_trx_split_line.budget_id = @budget_id
  AND expense_share_trx_split_line.expense_share_id = @expense_share_id
  AND expense_share_trx_split_line.budget_id IN (
    SELECT b.id
    FROM budget AS b
    WHERE b.id = @budget_id
      AND b.login_id = @login_id
  )
RETURNING *;

-- name: DeleteSplitLine :exec
DELETE FROM expense_share_trx_split_line
WHERE expense_share_trx_split_line.id = @id
  AND expense_share_trx_split_line.budget_id = @budget_id
  AND expense_share_trx_split_line.expense_share_id = @expense_share_id
  AND expense_share_trx_split_line.budget_id IN (
    SELECT b.id
    FROM budget AS b
    WHERE b.id = @budget_id
      AND b.login_id = @login_id
  );

-- name: GetShareMemberByBudget :one
SELECT
  sqlc.embed(bes),
  sqlc.embed(es),
  sqlc.embed(b)
FROM budget_expense_share AS bes
JOIN budget AS b ON bes.budget_id = b.id
JOIN expense_share AS es ON es.id = bes.expense_share_id
WHERE bes.expense_share_id = @expense_share_id
  AND bes.budget_id = @budget_id;
