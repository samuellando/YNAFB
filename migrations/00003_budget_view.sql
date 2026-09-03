-- +goose up

CREATE VIEW budget_month_categories AS
WITH one AS (
SELECT
  c.budget AS budget_id,
  m.month AS month,
  c.id AS category_id,
  CAST(COALESCE(a.allocated, 0) AS INTEGER) AS allocated,
  CAST(COALESCE(s.spent, 0) AS INTEGER) AS spent
FROM category AS c
JOIN (
  SELECT DISTINCT CAST(unixepoch(date(month, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER) AS month
  FROM allocation
  UNION
  SELECT DISTINCT CAST(unixepoch(date(t.date, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER) AS month
  FROM "transaction" AS t
  JOIN transaction_category AS tc ON tc."transaction" = t.id
  WHERE tc.category IS NOT NULL
) AS m
LEFT JOIN (
  SELECT a.category,
         CAST(unixepoch(date(a.month, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER) AS month,
         SUM(a.amount) AS allocated
  FROM allocation AS a
  GROUP BY a.category, month
) AS a ON a.category = c.id AND a.month = m.month
LEFT JOIN (
  SELECT ts.category,
         CAST(unixepoch(date(t.date, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER) AS month,
         SUM(ts.outflow - ts.inflow) AS spent
  FROM transaction_category AS ts
  JOIN "transaction" AS t ON t.id = ts."transaction"
  WHERE ts.category IS NOT NULL
  GROUP BY ts.category, month
) AS s ON s.category = c.id AND s.month = m.month)
SELECT
  budget_id,
  month,
  category_id,
  allocated,
  spent,
  SUM(MAX(allocated - spent, 0)) OVER (
    PARTITION by category_id
    ORDER BY month
  ) - MAX(allocated - spent, 0) + allocated - spent as available
FROM one
GROUP BY month, category_id
ORDER BY month, category_id;

-- +goose down

DROP VIEW budget_month_categories;
