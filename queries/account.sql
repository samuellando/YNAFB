-- name: CreateAccount :one
INSERT INTO account (
  budget,
  name
) VALUES (
  ?,
  ?
)
RETURNING *;
