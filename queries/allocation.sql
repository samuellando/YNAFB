-- name: CreateAllocation :one
INSERT INTO allocation (
  budget,
  category,
  amount
) VALUES (
  ?,
  ?,
  ?
)
RETURNING *;
