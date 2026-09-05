-- name: CreatePayee :one
INSERT INTO payee (
  budget_id,
  name
) VALUES (
  ?,
  ?
)
RETURNING *;

-- name: UpdatePayee :execrows
UPDATE payee
SET name = ?
WHERE id = ? AND budget_id = ?;

-- name: DeletePayee :exec
DELETE FROM payee
WHERE id = ? AND budget_id = ?;

-- name: ListPayees :many
SELECT id, budget_id, name
FROM payee
WHERE budget_id = ?
ORDER BY name;

-- name: GetPayeeByName :one
SELECT id, budget_id, name
FROM payee
WHERE budget_id = ? AND name = ?;