-- name: CreatePayeeDefaultSplit :one
INSERT INTO payee_default_split (
  payee,
  other_account,
  category,
  percent
) VALUES (
  ?,
  ?,
  ?,
  ?
)
RETURNING *;
