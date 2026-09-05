-- name: CreateTrxLine :one
INSERT INTO trx_line (
  budget_id,
  trx_id,
  dest_account_id,
  category_id,
  income,
  outflow,
  inflow
) VALUES (
  ?,
  ?,
  ?,
  ?,
  ?,
  ?,
  ?
)
RETURNING *;

-- name: UpdateTrxLine :execrows
UPDATE trx_line
SET trx_id = ?, dest_account_id = ?, category_id = ?, income = ?, outflow = ?, inflow = ?
WHERE id = ? AND budget_id = ?;

-- name: DeleteTrxLine :exec
DELETE FROM trx_line
WHERE id = ? AND budget_id = ?;

-- name: DeleteTrxLinesByTrx :exec
DELETE FROM trx_line
WHERE trx_id = ? AND budget_id = ?;