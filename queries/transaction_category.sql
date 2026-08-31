-- name: CreateTransactionCategory :one
INSERT INTO transaction_category (
  "transaction",
  other_account,
  category,
  income,
  outflow,
  inflow
) VALUES (
  ?,
  ?,
  ?,
  ?,
  ?,
  ?
)
RETURNING *;

-- name: UpdateTransactionCategory :execrows
UPDATE transaction_category
SET "transaction" = ?, other_account = ?, category = ?, income = ?, outflow = ?, inflow = ?
WHERE id = ?;
