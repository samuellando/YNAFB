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
