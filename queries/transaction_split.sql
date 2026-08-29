-- name: CreateTransactionSplit :one
INSERT INTO transaction_split (
  transaction,
  to_account,
  from_account,
  category,
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
