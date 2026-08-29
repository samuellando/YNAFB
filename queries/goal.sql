-- name: CreateGoal :one
INSERT INTO goal (
  budget,
  name,
  type,
  start,
  end,
  category,
  amount
) VALUES (
  ?,
  ?,
  ?,
  ?,
  ?,
  ?,
  ?
)
RETURNING *;
