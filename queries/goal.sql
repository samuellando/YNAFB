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

-- name: ListGoalsValues :many
WITH goal_data AS (
    SELECT
      g.id,
      g.type,
      g.category_id,
      g.start_date,
      g.end_date,
      g.amount,
      -- How much was allocated this month
      CAST(COALESCE(bmc.allocated, 0) AS INTEGER) AS allocated_mtd,
      -- Left over available from previous months
      CAST((CASE
          WHEN CAST(@month AS UNIX_EPOCH_INTEGER) < unixepoch() THEN
            COALESCE((
                SELECT MAX(prior.available, 0)
                FROM budget_month_categories AS prior
                WHERE prior.budget_id = g.budget_id AND prior.category_id = g.category_id AND prior.month < @month
                ORDER BY prior.month DESC
                LIMIT 1
            ), 0)
          ELSE 0
        END
      ) AS INTEGER) AS carry_over,
      CAST(MAX((strftime('%Y', g.end_date, 'unixepoch') - strftime('%Y', @month, 'unixepoch')) * 12
        + (strftime('%m', g.end_date, 'unixepoch') - strftime('%m', @month, 'unixepoch')), 0) AS INTEGER) as remaining_months,
      CAST(COALESCE((SELECT
          sum(allocated)
          FROM budget_month_categories AS bmcsum
          WHERE bmcsum.budget_id = g.budget_id AND bmcsum.month < @month AND bmcsum.month >= g.start_date AND bmcsum.category_id = g.category_id
      ), 0) AS INTEGER) as allocated
    FROM goal AS g
    LEFT JOIN budget_month_categories AS bmc ON bmc.month = @month AND bmc.category_id = g.category_id AND bmc.budget_id = g.budget_id
    JOIN budget AS b ON g.budget_id = b.id
    WHERE g.budget_id = @id AND b.login_id = @login_id AND g.start_date <= @month AND (g.end_date > @month OR g.end_date IS NULL)
), goal_need AS (
    SELECT
      *,
      CAST((
        CASE
          WHEN goal_data.type = 'monthly' THEN goal_data.amount
          WHEN goal_data.type = 'refill' THEN goal_data.amount - MAX(goal_data.carry_over, 0)
          WHEN goal_data.type = 'save' THEN (goal_data.amount - goal_data.allocated) / goal_data.remaining_months
        END
      ) AS INTEGER) AS amount_for_month
    FROM goal_data
)
SELECT
  goal_need.id,
  goal_need.type,
  goal_need.category_id,
  goal_need.start_date,
  goal_need.end_date,
  goal_need.amount,
  goal_need.allocated,
  goal_need.amount_for_month,
  CAST((goal_need.allocated_mtd - goal_need.amount_for_month) AS INTEGER) as gap
FROM goal_need;
