-- name: CreateTransactionSplit :one
INSERT INTO transaction_split (
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
