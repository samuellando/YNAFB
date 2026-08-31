-- name: CreateGoal :one
INSERT INTO goal (
  budget,
  type,
  start,
  "end",
  category,
  amount
) VALUES (
  ?,
  ?,
  ?,
  ?,
  ?,
  ?
)
RETURNING *;

-- name: UpdateGoal :execrows
UPDATE goal
SET type = ?, start = ?, "end" = ?, amount = ?
WHERE budget = ? AND category = ?;
