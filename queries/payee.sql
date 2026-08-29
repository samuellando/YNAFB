-- name: CreatePayee :one
INSERT INTO payee (
  budget,
  name
) VALUES (
  ?,
  ?
)
RETURNING *;
