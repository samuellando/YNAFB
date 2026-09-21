-- name: CreateGoal :one
INSERT INTO
  goal (budget_id, type, start_date, end_date, category_id, amount)
SELECT
  b.id, ?, ?, ?, ?, ?
FROM
  budget AS b
WHERE
  b.id = @budget_id AND b.login_id = @login_id
RETURNING
  *;

-- name: UpdateGoal :one
UPDATE goal
SET
  type = ?, start_date = ?, end_date = ?, amount = ?
WHERE
  goal.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id AND b.login_id = @login_id
  )
  AND goal.category_id = @category_id
RETURNING *;

-- name: DeleteGoal :exec
DELETE FROM goal
WHERE
  goal.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id AND b.login_id = @login_id
  )
  AND goal.category_id = @category_id;

-- name: ListGoals :many
SELECT g.*
FROM goal AS g
JOIN budget AS b ON g.budget_id = b.id
WHERE b.login_id = @login_id AND g.budget_id = @budget_id;

-- name: GetGoalByCategory :one
SELECT
  g.id,
  g.budget_id,
  g.type,
  g.start_date,
  g.end_date,
  g.category_id,
  g.amount
FROM goal AS g
JOIN budget AS b ON g.budget_id = b.id
WHERE b.login_id = @login_id AND g.budget_id = @budget_id AND g.category_id = @category_id;
