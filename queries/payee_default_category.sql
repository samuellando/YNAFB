-- name: CreatePayeeDefaultCategory :one
INSERT INTO payee_default_category (
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
