-- name: CreateBudget :one
INSERT INTO budget (
  name
) VALUES (
  ?
)
RETURNING *;

-- name: UpdateBudget :execrows
UPDATE budget
SET name = ?
WHERE id = ?;

-- name: DeleteBudget :exec
DELETE FROM budget
WHERE id = ?;

-- name: ListBudgets :many
SELECT id, name
FROM budget
ORDER BY id;

-- name: GetBudgetByName :one
SELECT id, name
FROM budget
WHERE name = ?;

-- name: ListBudgetActivityMonths :many
SELECT DISTINCT
  CAST(unixepoch(date(t.date, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER) as month
FROM "transaction" as t
JOIN account as acc ON t.account = acc.id
WHERE acc.budget = @budget
UNION
SELECT DISTINCT
  CAST(unixepoch(date(a.month, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER) as month
FROM allocation as a
WHERE a.budget = @budget
ORDER BY month;

-- name: ListBudgetMonthCategories :many
SELECT 
  c.id,
  c.id AS category_id,
  c.name AS category_name,
  cg.name AS category_gorup_name,
  COALESCE(allocated, 0) AS allocated,
  COALESCE(spent, 0) AS spent,
  CAST(CASE
    WHEN bmc.month IS NULL AND CAST(@month AS UNIX_EPOCH_INTEGER) < unixepoch() THEN COALESCE((
      SELECT MAX(prior.available, 0)
      FROM budget_month_categories AS prior
      WHERE prior.category_id = c.id AND prior.month < @month
      ORDER BY prior.month DESC
      LIMIT 1
    ), 0)
    ELSE COALESCE(bmc.available, 0)
  END AS INTEGER) AS available
FROM category AS c
LEFT JOIN category_group AS cg ON c.category_group = cg.id
LEFT JOIN budget_month_categories AS bmc ON bmc.category_id = c.id AND bmc.month = @month
WHERE c.budget = @budget 
ORDER BY cg.name ASC NULLS FIRST, c.name;

-- name: GetBudgetMonthSummary :one
SELECT
  COALESCE(bms.ready_to_assign, (
    SELECT prev.ready_to_assign
    FROM budget_month_summary AS prev
    WHERE prev.month < @month
    ORDER BY prev.month desc
    LIMIT 1
  ) , 0) AS ready_to_assign,
  COALESCE(bms.income, 0) AS income,
  COALESCE(bms.allocated, 0) AS allocated,
  COALESCE(bms.spent, 0) AS spent,
  COALESCE(bms.available, (
    SELECT prev.available
    FROM budget_month_summary AS prev
    WHERE prev.month < @month
    ORDER BY prev.month desc
    LIMIT 1
  ) , 0) AS available,
  COALESCE(bms.uncategorized, 0) AS uncategorized
FROM (SELECT @month AS month) AS requested
LEFT JOIN budget_month_summary AS bms ON bms.month = requested.month;
