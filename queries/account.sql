-- name: CreateAccount :one
INSERT INTO account (
  budget,
  name
) VALUES (
  ?,
  ?
)
RETURNING *;

-- name: UpdateAccount :execrows
UPDATE account
SET name = ?
WHERE id = ?;
