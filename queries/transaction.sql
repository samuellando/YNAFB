-- name: CreateTransaction :one
INSERT INTO transaction (
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
