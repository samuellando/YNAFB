-- name: CreateTrx :one
INSERT INTO trx (
  budget_id,
  account_id,
  payee_id,
  date,
  total_outflow,
  total_inflow,
  note
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

-- name: UpdateTrx :execrows
UPDATE trx
SET date = ?, account_id = ?, payee_id = ?, total_outflow = ?, total_inflow = ?, note = ?
WHERE id = ? AND budget_id = ?;

-- name: DeleteTrx :exec
DELETE FROM trx
WHERE id = ? AND budget_id = ?;

-- name: ListTrxs :many
SELECT
  t.id,
  t.budget_id,
  t.date,
  a.name AS account_name,
  p.name AS payee_name,
  t.total_outflow,
  t.total_inflow,
  t.note
FROM trx AS t
JOIN account AS a ON a.id = t.account_id
JOIN payee AS p ON p.id = t.payee_id
WHERE t.budget_id = @budget_id
ORDER BY t.date DESC, t.id DESC;