-- name: CreateBudget :one
INSERT INTO budget (
  login_id,
  name
) VALUES (
  ?,
  ?
)
RETURNING *;

-- name: UpdateBudget :execrows
UPDATE budget
SET
  name = ?
WHERE
  budget.login_id = @login_id AND budget.id = @id;

-- name: DeleteBudget :exec
DELETE FROM budget
WHERE
  budget.login_id = @login_id AND budget.id = @id;

-- name: ListBudgets :many
SELECT
  budget.id,
  budget.login_id,
  budget.name
FROM
  budget
WHERE
  budget.login_id = @login_id
ORDER BY
  budget.id;

-- name: GetBudgetByName :one
SELECT
  budget.id,
  budget.login_id,
  budget.name
FROM
  budget
WHERE
  budget.login_id = @login_id AND budget.name = ?;

-- name: ListBudgetMonthCategories :many
SELECT
  c.id,
  c.id AS category_id,
  c.budget_id,
  c.name AS category_name,
  cg.name AS category_group_name,
  COALESCE(bmc.allocated, 0) AS allocated,
  COALESCE(bmc.spent, 0) AS spent,
  CAST(CASE
    WHEN bmc.month IS NULL AND CAST(@month AS UNIX_EPOCH_INTEGER) < unixepoch() THEN COALESCE((
      SELECT MAX(prior.available, 0)
      FROM budget_month_categories AS prior
      WHERE prior.category_id = c.id AND prior.budget_id = c.budget_id AND prior.month < @month
      ORDER BY prior.month DESC
      LIMIT 1
    ), 0)
    ELSE COALESCE(bmc.available, 0)
  END AS INTEGER) AS available
FROM category AS c
JOIN budget AS b ON c.budget_id = b.id
LEFT JOIN category_group AS cg ON c.category_group_id = cg.id
LEFT JOIN budget_month_categories AS bmc ON bmc.category_id = c.id AND bmc.month = @month AND bmc.budget_id = c.budget_id
WHERE b.login_id = @login_id AND c.budget_id = @id
ORDER BY cg.name ASC NULLS FIRST, c.name;

-- name: GetBudgetMonthSummary :one
SELECT
  COALESCE(bms.ready_to_assign, (
    SELECT prev.ready_to_assign
    FROM budget_month_summary AS prev
    WHERE prev.budget_id = @id AND prev.month < @month
    ORDER BY prev.month desc
    LIMIT 1
  ) , 0) AS ready_to_assign,
  COALESCE(bms.income, 0) AS income,
  COALESCE(bms.allocated, 0) AS allocated,
  COALESCE(bms.spent, 0) AS spent,
  COALESCE(bms.available, (
    SELECT prev.available
    FROM budget_month_summary AS prev
    WHERE prev.budget_id = @id AND prev.month < @month
    ORDER BY prev.month desc
    LIMIT 1
  ) , 0) AS available,
  COALESCE(bms.uncategorized, 0) AS uncategorized
FROM (SELECT @month AS month) AS requested
LEFT JOIN budget_month_summary AS bms ON bms.month = requested.month AND bms.budget_id = @id
JOIN budget AS b ON b.id = @id
WHERE b.login_id = @login_id;
