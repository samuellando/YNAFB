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
