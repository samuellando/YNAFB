-- name: CreatePayeeDefaultLine :one
INSERT INTO
  payee_default_line (
    budget_id,
    payee_id,
    dest_account_id,
    category_id,
    expense_share_id,
    split_budget_expense_share_id,
    dest_budget_expense_share_id,
    income,
    percent
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
  ?
FROM
  budget AS b
WHERE
  b.id = @budget_id
  AND b.login_id = @login_id
RETURNING
  *;

-- name: UpdatePayeeDefaultLine :one
UPDATE payee_default_line
SET
  payee_id = ?,
  dest_account_id = ?,
  category_id = ?,
  expense_share_id = ?,
  split_budget_expense_share_id = ?,
  dest_budget_expense_share_id = ?,
  income = ?,
  percent = ?
WHERE
  payee_default_line.id = @id
  AND payee_default_line.budget_id IN (
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

-- name: DeletePayeeDefaultLine :exec
DELETE FROM payee_default_line
WHERE
  payee_default_line.id = @id
  AND payee_default_line.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id
      AND b.login_id = @login_id
  );

-- name: GetPayeeDefaultLine :one
SELECT
  pdl.*,
  c.name AS category_name,
  cg.id AS category_group_id,
  cg.name AS category_group_name,
  a.name AS dest_account_name,
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
  payee_default_line AS pdl
  JOIN budget AS b ON pdl.budget_id = b.id
  LEFT JOIN category AS c ON c.id = pdl.category_id
  LEFT JOIN category_group AS cg ON c.category_group_id = cg.id
  LEFT JOIN account AS a ON a.id = pdl.dest_account_id
  LEFT JOIN expense_share AS es ON es.id = pdl.expense_share_id
  LEFT JOIN budget_expense_share AS sbes ON sbes.id = pdl.split_budget_expense_share_id
  LEFT JOIN budget AS sb ON sb.id = sbes.budget_id
  LEFT JOIN budget_expense_share AS dbes ON dbes.id = pdl.dest_budget_expense_share_id
  LEFT JOIN budget AS db ON db.id = dbes.budget_id
WHERE
  b.login_id = @login_id
  AND pdl.budget_id = @budget_id
  AND pdl.id = @id;

-- name: ListPayeeDefaultLinesByPayee :many
SELECT
  pdl.id,
  pdl.budget_id,
  pdl.payee_id,
  pdl.dest_account_id,
  ao.name AS dest_account_name,
  pdl.category_id,
  c.name AS category_name,
  cg.id AS category_group_id,
  cg.name AS category_group_name,
  pdl.income,
  pdl.percent,
  es.default_name AS expense_share_default_name,
  pdl.expense_share_id,
  pdl.split_budget_expense_share_id,
  pdl.dest_budget_expense_share_id,
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
  payee_default_line AS pdl
  JOIN budget AS b ON pdl.budget_id = b.id
  LEFT JOIN account AS ao ON ao.id = pdl.dest_account_id
  LEFT JOIN category AS c ON c.id = pdl.category_id
  LEFT JOIN category_group AS cg ON c.category_group_id = cg.id
  LEFT JOIN expense_share AS es ON es.id = pdl.expense_share_id
  LEFT JOIN budget_expense_share AS sbes ON sbes.id = pdl.split_budget_expense_share_id
  LEFT JOIN budget AS sb ON sb.id = sbes.budget_id
  LEFT JOIN budget_expense_share AS dbes ON dbes.id = pdl.dest_budget_expense_share_id
  LEFT JOIN budget AS db ON db.id = dbes.budget_id
WHERE
  b.login_id = @login_id
  AND pdl.payee_id = @payee_id
  AND pdl.budget_id = @budget_id
ORDER BY
  pdl.id;
