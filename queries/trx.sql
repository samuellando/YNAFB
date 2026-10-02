-- name: CreateTrx :one
INSERT INTO
  trx (
    budget_id,
    account_id,
    payee_id,
    date,
    total_outflow,
    total_inflow,
    note
  )
SELECT
  b.id,
  ?,
  ?,
  ?,
  ?,
  ?,
  ?
FROM
  budget AS b
WHERE
  b.id = @budget_id
  AND b.login_id = @login_id
RETURNING
  *;

-- name: UpdateTrx :one
UPDATE trx
SET
  date = ?,
  payee_id = ?,
  total_outflow = ?,
  total_inflow = ?,
  note = ?
WHERE
  trx.id = @id
  AND trx.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id
      AND b.login_id = @login_id
  )
RETURNING
  *;

-- name: DeleteTrx :exec
DELETE FROM trx
WHERE
  trx.id = @id
  AND trx.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id
      AND b.login_id = @login_id
  );

-- name: GetTrxAndLines :many
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
  sb.id AS split_budget_id,
  sb.name AS split_budget_name,
  sb.login_id AS split_budget_login_id,
  db.id AS dest_budget_id,
  db.name AS dest_budget_name,
  db.login_id AS dest_budget_login_id,
  CAST((r.account_id IS NOT NULL) AS bool) AS reconciled
FROM
  trx AS t
  JOIN budget AS b ON t.budget_id = b.id
  JOIN account AS a ON t.account_id = a.id
  JOIN payee AS p ON p.id = t.payee_id
  LEFT JOIN trx_line AS tl ON t.id = tl.trx_id
  LEFT JOIN category AS c ON tl.category_id = c.id
  LEFT JOIN category_group AS cg ON c.category_group_id = cg.id
  LEFT JOIN account AS da ON tl.dest_account_id = da.id
  LEFT JOIN expense_share AS es ON es.id = tl.expense_share_id
  LEFT JOIN budget_expense_share AS sbes ON sbes.id = tl.split_budget_expense_share_id
  LEFT JOIN budget AS sb ON sb.id = sbes.budget_id
  LEFT JOIN budget_expense_share AS dbes ON dbes.id = tl.dest_budget_expense_share_id
  LEFT JOIN budget AS db ON db.id = dbes.budget_id
  LEFT JOIN reconciliation AS r ON r.account_id = a.id
  AND r.trx_id = t.id
WHERE
  b.login_id = @login_id
  AND t.budget_id = @budget_id
  AND t.id = @id
  AND a.id = @account_id
ORDER BY
  t.date DESC,
  t.id DESC;

-- name: ListTrxsAndLines :many
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
  CAST((r.account_id IS NOT NULL) AS bool) AS reconciled
FROM
  trx AS t
  JOIN budget AS b ON t.budget_id = b.id
  JOIN account AS a ON t.account_id = a.id
  JOIN payee AS p ON p.id = t.payee_id
  LEFT JOIN trx_line AS tl ON t.id = tl.trx_id
  LEFT JOIN category AS c ON tl.category_id = c.id
  LEFT JOIN category_group AS cg ON c.category_group_id = cg.id
  LEFT JOIN account AS da ON tl.dest_account_id = da.id
  LEFT JOIN expense_share AS es ON es.id = tl.expense_share_id
  LEFT JOIN budget_expense_share AS sbes ON sbes.id = tl.split_budget_expense_share_id
  LEFT JOIN budget AS sb ON sb.id = sbes.budget_id
  LEFT JOIN budget_expense_share AS dbes ON dbes.id = tl.dest_budget_expense_share_id
  LEFT JOIN budget AS db ON db.id = dbes.budget_id
  LEFT JOIN reconciliation AS r ON r.account_id = a.id
  AND r.trx_id = t.id
WHERE
  b.login_id = @login_id
  AND t.budget_id = @budget_id
ORDER BY
  t.date DESC,
  t.id DESC;
