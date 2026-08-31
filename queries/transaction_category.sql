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
