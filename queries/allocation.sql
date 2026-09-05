-- name: CreateAllocation :one
INSERT INTO allocation (
  budget_id,
  category_id,
  month,
  amount
) VALUES (
  ?,
  ?,
  ?,
  ?
)
RETURNING *;

-- name: UpdateAllocation :execrows
UPDATE allocation
SET amount = ?
WHERE budget_id = ? AND category_id = ? AND month = ?;

-- name: DeleteAllocationByCategoryAndMonth :exec
DELETE FROM allocation
WHERE budget_id = ? AND category_id = ? AND month = ?;

-- name: ListAllocations :many
SELECT
  a.id,
  a.budget_id,
  a.month,
  a.category_id,
  c.name AS category_name,
  a.amount
FROM allocation AS a
JOIN category AS c ON c.id = a.category_id
WHERE a.budget_id = @budget_id
ORDER BY a.month, c.name;