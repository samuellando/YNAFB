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
  COALESCE(a.amount, 0) AS allocated,
  CAST(COALESCE(s.spent, 0) AS INTEGER) AS spent
FROM category AS c
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
ORDER BY c.name;
