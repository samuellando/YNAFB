-- name: CreateGoal :one
INSERT INTO goal (
  budget_id,
  type,
  start_date,
  end_date,
  category_id,
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
SET type = ?, start_date = ?, end_date = ?, amount = ?
WHERE budget_id = ? AND category_id = ?;

-- name: DeleteGoal :exec
DELETE FROM goal
WHERE id = ? AND budget_id = ?;

-- name: ListGoals :many
SELECT
  g.id,
  g.budget_id,
  g.type,
  g.start_date,
  g.end_date,
  g.category_id,
  c.name AS category_name,
  g.amount
FROM goal AS g
JOIN category AS c ON c.id = g.category_id
WHERE g.budget_id = @budget_id
ORDER BY c.name;

-- name: GetGoalByCategory :one
SELECT id, budget_id, type, start_date, end_date, category_id, amount
FROM goal
WHERE budget_id = ? AND category_id = ?;

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
      CAST(MAX((strftime('%Y', end_date, 'unixepoch') - strftime('%Y', @month, 'unixepoch')) * 12 
        + (strftime('%m', end_date, 'unixepoch') - strftime('%m', @month, 'unixepoch')), 0) AS INTEGER) as remaining_months,
      CAST(COALESCE((SELECT 
          sum(allocated)
          FROM budget_month_categories AS bmcsum
          WHERE bmcsum.budget_id = g.budget_id AND bmcsum.month < @month AND bmcsum.month >= start_date AND bmcsum.category_id = g.category_id
      ), 0) AS INTEGER) as allocated
    FROM goal AS g
    LEFT JOIN budget_month_categories AS bmc ON bmc.month = @month AND bmc.category_id = g.category_id AND bmc.budget_id = g.budget_id
    WHERE g.budget_id = @budget_id AND start_date <= @month AND (end_date > @month OR end_date IS NULL)
), goal_need AS (
    SELECT
      *,
      CAST((
        CASE
          WHEN type = 'monthly' THEN amount
          WHEN type = 'refill' THEN amount - MAX(carry_over, 0)
          WHEN type = 'save' THEN (amount - allocated) / remaining_months
        END
      ) AS INTEGER) AS amount_for_month 
    FROM goal_data
)
SELECT 
id,
type,
category_id,
start_date,
end_date,
amount,
allocated,
amount_for_month,
  CAST((allocated_mtd - amount_for_month) AS INTEGER) as gap
FROM goal_need;