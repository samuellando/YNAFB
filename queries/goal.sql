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

-- name: GetGoalByCategory :one
SELECT id, budget, type, start, "end", category, amount
FROM goal
WHERE budget = ? AND category = ?;

-- name: ListGoalsValues :many
WITH goal_data AS (
    SELECT 
      id,
      type,
      category,
      start,
      "end",
      amount,
      -- How much was allocated this month
      CAST(COALESCE(allocated, 0) AS INTEGER) AS allocated_mtd,
      -- Left over available from previous months
      CAST((CASE
          WHEN CAST(@month AS UNIX_EPOCH_INTEGER) < unixepoch() THEN
            COALESCE((
                SELECT MAX(prior.available, 0)
                FROM budget_month_categories AS prior
                WHERE prior.category_id = g.category AND prior.month < @month
                ORDER BY prior.month DESC
                LIMIT 1
            ), 0)
          ELSE 0
        END
      ) AS INTEGER) AS carry_over,
      CAST(MAX((strftime('%Y', "end", 'unixepoch') - strftime('%Y', @month, 'unixepoch')) * 12 
        + (strftime('%m', "end", 'unixepoch') - strftime('%m', @month, 'unixepoch')), 0) AS INTEGER) as remaining_months,
      CAST(COALESCE((SELECT 
          sum(allocated)
          FROM budget_month_categories AS bmcsum
          WHERE bmcsum.month < @month AND bmcsum.month >= start AND category_id = g.category
      ), 0) AS INTEGER) as allocated
    FROM goal AS g
    LEFT JOIN budget_month_categories AS bmc ON bmc.month = @month AND bmc.category_id = g.category
    WHERE budget = @budget AND start <= @month AND ("end" > @month OR "end" IS NULL)
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
category,
start,
"end",
amount,
allocated,
amount_for_month,
  CAST((allocated_mtd - amount_for_month) AS INTEGER) as gap
FROM goal_need
