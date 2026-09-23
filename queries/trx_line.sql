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
  dest_account_id = ?, category_id = ?, income = ?, outflow = ?, inflow = ?
WHERE
  trx_line.id = @id
  AND trx_line.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id AND b.login_id = @login_id AND trx_line.trx_id = @trx_id
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

-- name: GetTrxLine :one
SELECT
  sqlc.embed(tl),
  b.name AS budget_name,
  da.name AS dest_account_name,
  c.name AS category_name,
  cg.id AS category_group_id,
  cg.name AS category_group_name
FROM trx_line AS tl
JOIN budget AS b ON tl.budget_id = b.id
LEFT JOIN account AS da ON da.id = tl.dest_account_id
LEFT JOIN category AS c ON c.id = tl.category_id
LEFT JOIN category_group AS cg ON cg.id = c.category_group_id
WHERE
  b.login_id = @login_id
  AND tl.budget_id = @budget_id
  AND tl.id = @id
  AND tl.trx_id = @trx_id;

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
