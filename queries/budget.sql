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
  c.name AS category_name,
  cg.name AS category_gorup_name,
  COALESCE(allocated, 0) AS allocated,
  COALESCE(spent, 0) AS spend,
  COALESCE(available, 0) AS available
FROM category AS c
LEFT JOIN category_group AS cg ON c.category_group = cg.id
LEFT JOIN budget_month_categories AS bmc ON bmc.category_id = c.id AND bmc.month = @month
WHERE c.budget = @budget 
ORDER BY gc.name, c.name;

-- name: ListCategoryMonthlySpendingByBudget :many
SELECT
  ts.category,
  CAST(strftime('%Y', t.date, 'unixepoch') AS INTEGER) * 100 + CAST(strftime('%m', t.date, 'unixepoch') AS INTEGER) AS month,
  CAST(SUM(ts.outflow - ts.inflow) AS INTEGER) AS net
FROM transaction_category AS ts
JOIN "transaction" AS t ON t.id = ts."transaction"
JOIN category AS c ON c.id = ts.category
WHERE c.budget = @budget
  AND t.date < @end
GROUP BY ts.category, month;

-- name: GetBudgetBalanceAsOf :one
SELECT
  (SELECT COALESCE(SUM(t.total_inflow - t.total_outflow), 0)
     FROM "transaction" AS t
     JOIN account AS a ON a.id = t.account
    WHERE a.budget = @budget AND t.date < @end)
  + (SELECT COALESCE(SUM(ts.outflow - ts.inflow), 0)
       FROM transaction_category AS ts
       JOIN account AS a ON a.id = ts.other_account
       JOIN "transaction" AS t ON t.id = ts."transaction"
      WHERE a.budget = @budget AND t.date < @end) AS balance;

-- name: GetUncategorizedAmountByBudget :one
SELECT CAST(COALESCE(SUM(t.total_outflow + t.total_inflow), 0) AS INTEGER)
FROM "transaction" AS t
JOIN account AS a ON a.id = t.account
WHERE a.budget = @budget
  AND t.date < @end
  AND NOT EXISTS (
    SELECT 1 FROM transaction_category AS ts WHERE ts."transaction" = t.id
  );

-- name: GetIncomeByBudgetMonth :one
SELECT CAST(COALESCE(SUM(ts.inflow - ts.outflow), 0) AS INTEGER)
FROM transaction_category AS ts
JOIN "transaction" AS t ON t.id = ts."transaction"
JOIN account AS a ON a.id = t.account
WHERE a.budget = @budget
  AND ts.income
  AND t.date >= @start AND t.date < @end;
