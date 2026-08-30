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

-- name: ListAccountTransactions :many
SELECT
  t.id,
  t.date,
  t.account AS transaction_account,
  ta.name AS transaction_account_name,
  p.name AS payee_name,
  t.total_outflow,
  t.total_inflow,
  EXISTS (
    SELECT 1 FROM reconciliation AS r
    WHERE r.account = @account AND r."transaction" = t.id
  ) AS reconciled,
  t.note,
  ts.id AS category_id,
  ts.other_account,
  ao.name AS other_account_name,
  ts.category,
  c.name AS category_name,
  ts.outflow,
  ts.inflow
FROM "transaction" AS t
JOIN account AS ta ON ta.id = t.account
JOIN payee AS p ON p.id = t.payee
LEFT JOIN transaction_category AS ts ON ts."transaction" = t.id
LEFT JOIN account AS ao ON ao.id = ts.other_account
LEFT JOIN category AS c ON c.id = ts.category
WHERE t.account = @account OR ts.other_account = @other_account
ORDER BY t.date DESC, payee ASC, t.id ASC, ts.id ASC;
