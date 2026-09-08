-- name: CreateTrxLine :one
INSERT INTO
  trx_line (budget_id, trx_id, dest_account_id, category_id, income, outflow, inflow)
SELECT
  b.id, ?, ?, ?, ?, ?, ?
FROM
  budget AS b
WHERE
  b.id = @budget_id AND b.login_id = @login_id
RETURNING
  *;

-- name: UpdateTrxLine :one
UPDATE trx_line
SET
  trx_id = ?, dest_account_id = ?, category_id = ?, income = ?, outflow = ?, inflow = ?
WHERE
  trx_line.id = @id
  AND trx_line.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id AND b.login_id = @login_id
  )
RETURNING *;

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
      b.id = @budget_id AND b.login_id = @login_id
  );

-- name: DeleteTrxLinesByTrx :exec
DELETE FROM trx_line
WHERE
  trx_line.trx_id = @trx_id
  AND trx_line.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id AND b.login_id = @login_id
  );