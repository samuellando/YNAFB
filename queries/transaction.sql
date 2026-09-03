-- name: CreateTransaction :one
INSERT INTO "transaction" (
  date,
  account,
  payee,
  total_outflow,
  total_inflow,
  note
) VALUES (
  ?,
  ?,
  ?,
  ?,
  ?,
  ?
)
RETURNING *;

-- name: UpdateTransaction :execrows
UPDATE "transaction"
SET date = ?, account = ?, payee = ?, total_outflow = ?, total_inflow = ?, note = ?
WHERE id = ?;

-- name: DeleteTransaction :exec
DELETE FROM "transaction"
WHERE id = ?;

-- name: ListTransactions :many
SELECT
  t.id,
  t.date,
  a.name AS account_name,
  p.name AS payee_name,
  t.total_outflow,
  t.total_inflow,
  t.note
FROM "transaction" AS t
JOIN account AS a ON a.id = t.account
JOIN payee AS p ON p.id = t.payee
WHERE a.budget = ?
ORDER BY t.date DESC, t.id DESC;
