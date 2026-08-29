-- name: CreatePayeeDefaultSplit :one
INSERT INTO payee_default_split (
  payee,
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
