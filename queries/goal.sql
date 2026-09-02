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

-- name: DeleteGoal :exec
DELETE FROM goal
WHERE id = ?;

-- name: ListGoals :many
SELECT
  g.id,
  g.type,
  g.start,
  g."end",
  g.category AS category_id,
  c.name AS category_name,
  g.amount
FROM goal AS g
JOIN category AS c ON c.id = g.category
WHERE g.budget = ?
ORDER BY c.name;
