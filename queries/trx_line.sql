-- name: CreateTrxLine :one
INSERT INTO
  trx_line (
    budget_id,
    trx_id,
    dest_account_id,
    category_id,
    income,
    expense_share_id,
    split_budget_expense_share_id,
    dest_budget_expense_share_id,
    outflow,
    inflow
  )
SELECT
  b.id,
  ?,
  ?,
  ?,
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

-- name: UpdateTrxLine :one
UPDATE trx_line
SET
  dest_account_id = ?,
  category_id = ?,
  income = ?,
  outflow = ?,
  inflow = ?,
  expense_share_id = ?,
  split_budget_expense_share_id = ?,
  dest_budget_expense_share_id = ?
WHERE
  trx_line.id = @id
  AND trx_line.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id
      AND b.login_id = @login_id
      AND trx_line.trx_id = @trx_id
  )
RETURNING
  *;

-- name: DeleteTrxLine :exec
DELETE FROM trx_line
WHERE
  trx_line.id = @id
  AND trx_line.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id
      AND b.login_id = @login_id
  );

-- name: GetTrxLine :one
SELECT
  sqlc.embed(tl),
  b.name AS budget_name,
  da.name AS dest_account_name,
  c.name AS category_name,
  cg.id AS category_group_id,
  cg.name AS category_group_name,
  es.default_name AS expense_share_default_name,
  sbes.name AS split_budget_expense_share_name,
  sbes.display_name AS split_budget_expense_share_display_name,
  sb.id AS split_budget_id,
  sb.name AS split_budget_name,
  sb.login_id AS split_budget_login_id,
  dbes.name AS dest_budget_expense_share_name,
  dbes.display_name AS dest_budget_expense_share_display_name,
  db.id AS dest_budget_id,
  db.name AS dest_budget_name,
  db.login_id AS dest_budget_login_id
FROM
  trx_line AS tl
  JOIN budget AS b ON tl.budget_id = b.id
  LEFT JOIN account AS da ON da.id = tl.dest_account_id
  LEFT JOIN category AS c ON c.id = tl.category_id
  LEFT JOIN category_group AS cg ON cg.id = c.category_group_id
  LEFT JOIN expense_share AS es ON es.id = tl.expense_share_id
  LEFT JOIN budget_expense_share AS sbes ON sbes.id = tl.split_budget_expense_share_id
  LEFT JOIN budget AS sb ON sb.id = sbes.budget_id
  LEFT JOIN budget_expense_share AS dbes ON dbes.id = tl.dest_budget_expense_share_id
  LEFT JOIN budget AS db ON db.id = dbes.budget_id
WHERE
  b.login_id = @login_id
  AND tl.budget_id = @budget_id
  AND tl.id = @id
  AND tl.trx_id = @trx_id;
