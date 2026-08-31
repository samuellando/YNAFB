-- name: CreatePayee :one
INSERT INTO payee (
  budget,
  name
) VALUES (
  ?,
  ?
)
RETURNING *;

-- name: UpdatePayee :execrows
UPDATE payee
SET name = ?
WHERE id = ?;
