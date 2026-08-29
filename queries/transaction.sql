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
  t.reconciled,
  t.note,
  ts.id AS split_id,
  ts.to_account,
  ato.name AS to_account_name,
  ts.from_account,
  afrom.name AS from_account_name,
  ts.category,
  c.name AS category_name,
  ts.outflow,
  ts.inflow
FROM "transaction" AS t
JOIN payee AS p ON p.id = t.payee
JOIN transaction_split AS ts ON ts."transaction" = t.id
LEFT JOIN account AS ato ON ato.id = ts.to_account
LEFT JOIN account AS afrom ON afrom.id = ts.from_account
LEFT JOIN category AS c ON c.id = ts.category
WHERE t.account = ?
ORDER BY t.date ASC, t.id ASC, ts.id ASC;
