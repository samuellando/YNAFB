-- name: CreateTransaction :one
INSERT INTO "transaction" (
  date,
  account,
  payee,
  total_outflow,
  total_inflow,
  reconciled,
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

-- name: ListAccountTransactions :many
SELECT
  t.id,
  t.date,
  p.name AS payee_name,
  t.total_outflow,
  t.total_inflow,
  t.reconciled,
  t.note,
  ts.id AS split_id,
  ts.other_account,
  ao.name AS other_account_name,
  ts.category,
  c.name AS category_name,
  ts.outflow,
  ts.inflow
FROM "transaction" AS t
JOIN payee AS p ON p.id = t.payee
LEFT JOIN transaction_split AS ts ON ts."transaction" = t.id
LEFT JOIN account AS ao ON ao.id = ts.other_account
LEFT JOIN category AS c ON c.id = ts.category
WHERE t.account = ?
ORDER BY t.date ASC, t.id ASC, ts.id ASC;
