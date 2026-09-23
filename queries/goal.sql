-- name: CreateGoal :one
INSERT INTO
  goal (
    budget_id,
    type,
    start_date,
    end_date,
    category_id,
    amount
  )
SELECT
  b.id,
  ?,
  ?,
  ?,
  ?,
  ?
FROM
  budget AS b
WHERE
  b.id = @budget_id
  AND b.login_id = @login_id
RETURNING
  *;

-- name: UpdateGoal :one
UPDATE goal
SET
  type = ?,
  start_date = ?,
  end_date = ?,
  amount = ?
WHERE
  goal.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id
      AND b.login_id = @login_id
  )
  AND goal.category_id = @category_id
RETURNING
  *;

-- name: DeleteGoal :exec
DELETE FROM goal
WHERE
  goal.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id
      AND b.login_id = @login_id
  )
  AND goal.category_id = @category_id;

-- name: ListGoals :many
SELECT
  g.*,
  c.name AS category_name,
  cg.id AS category_group_id,
  cg.name AS category_group_name
FROM
  goal AS g
  JOIN budget AS b ON g.budget_id = b.id
  JOIN category AS c ON g.category_id = c.id
  LEFT JOIN category_group AS cg ON cg.id = c.category_group_id
WHERE
  b.login_id = @login_id
  AND g.budget_id = @budget_id;

-- name: GetGoalByCategory :one
SELECT
  g.id,
  g.budget_id,
  g.type,
  g.start_date,
  g.end_date,
  g.amount,
  g.category_id,
  c.name AS category_name,
  cg.id AS category_group_id,
  cg.name AS category_group_name
FROM
  goal AS g
  JOIN budget AS b ON g.budget_id = b.id
  JOIN category AS c ON g.category_id = c.id
  LEFT JOIN category_group AS cg ON cg.id = c.category_group_id
WHERE
  b.login_id = @login_id
  AND g.budget_id = @budget_id
  AND g.category_id = @category_id;
