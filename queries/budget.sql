-- name: CreateBudget :one
INSERT INTO budget (
  name
) VALUES (
  ?
)
RETURNING *;

-- name: ListAllocationMonthsByBudget :many
SELECT DISTINCT month
FROM allocation
WHERE budget = ?
ORDER BY month;

-- name: ListTransactionDatesByBudget :many
SELECT t.date
FROM "transaction" AS t
JOIN account AS a ON a.id = t.account
WHERE a.budget = ?;

-- name: ListBudgetMonthCategories :many
SELECT
  c.id,
  c.name,
  g.name,
  COALESCE(a.amount, 0) AS allocated,
  CAST(COALESCE(s.spent, 0) AS INTEGER) AS spent
FROM category AS c
LEFT JOIN category_group AS g ON c.category_group = g.id
LEFT JOIN allocation AS a
  ON a.category = c.id
 AND a.budget = @budget
 AND a.month >= @start AND a.month < @end
LEFT JOIN (
  SELECT
    ts.category,
    SUM(ts.outflow - ts.inflow) AS spent
  FROM transaction_category AS ts
  JOIN "transaction" AS t ON t.id = ts."transaction"
  WHERE t.date >= @start AND t.date < @end
    AND ts.category IS NOT NULL
  GROUP BY ts.category
) AS s ON s.category = c.id
WHERE c.budget = @budget
ORDER BY g.name, c.name;

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
