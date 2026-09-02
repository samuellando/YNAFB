-- name: CreateAllocation :one
INSERT INTO allocation (
  budget,
  category,
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
WHERE budget = ? AND category = ? AND month = ?;

-- name: DeleteAllocationByCategoryAndMonth :exec
DELETE FROM allocation
WHERE budget = ? AND category = ? AND month = ?;

-- name: ListAllocations :many
SELECT
  a.id,
  a.month,
  a.category AS category_id,
  c.name AS category_name,
  a.amount
FROM allocation AS a
JOIN category AS c ON c.id = a.category
WHERE a.budget = ?
ORDER BY a.month, c.name;

