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

-- name: DeletePayee :exec
DELETE FROM payee
WHERE id = ?;

-- name: ListPayees :many
SELECT id, budget, name
FROM payee
WHERE budget = ?
ORDER BY name;
