-- name: CreatePayeeDefaultLine :one
INSERT INTO
  payee_default_line (budget_id, payee_id, dest_account_id, category_id, income, percent)
SELECT
  b.id, ?, ?, ?, ?, ?
FROM
  budget AS b
WHERE
  b.id = @budget_id AND b.login_id = @login_id
RETURNING
  *;

-- name: UpdatePayeeDefaultLine :one
UPDATE payee_default_line
SET
  payee_id = ?, dest_account_id = ?, category_id = ?, income = ?, percent = ?
WHERE
  payee_default_line.id = @id
  AND payee_default_line.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id AND b.login_id = @login_id
  )
RETURNING *;

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
      b.id = @budget_id AND b.login_id = @login_id
  );

-- name: DeletePayeeDefaultLinesByPayee :exec
DELETE FROM payee_default_line
WHERE
  payee_default_line.payee_id = @payee_id
  AND payee_default_line.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id AND b.login_id = @login_id
  );

-- name: ListPayeeDefaultLinesByPayee :many
SELECT
  pdl.id,
  pdl.budget_id,
  pdl.payee_id,
  pdl.dest_account_id,
  ao.name AS dest_account_name,
  pdl.category_id,
  c.name AS category_name,
  pdl.income,
  pdl.percent
FROM payee_default_line AS pdl
JOIN budget AS b ON pdl.budget_id = b.id
LEFT JOIN account AS ao ON ao.id = pdl.dest_account_id
LEFT JOIN category AS c ON c.id = pdl.category_id
WHERE b.login_id = @login_id AND pdl.payee_id = @payee_id AND pdl.budget_id = @budget_id
ORDER BY pdl.id;
