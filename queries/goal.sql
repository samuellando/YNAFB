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
