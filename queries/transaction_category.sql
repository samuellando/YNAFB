-- name: CreateTransactionCategory :one
INSERT INTO transaction_category (
  "transaction",
  other_account,
  category,
  outflow,
  inflow
) VALUES (
  ?,
  ?,
  ?,
  ?,
  ?
)
RETURNING *;
