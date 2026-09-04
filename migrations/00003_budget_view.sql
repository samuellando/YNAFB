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

CREATE VIEW budget_month_summary AS
WITH months AS (
  SELECT DISTINCT 
    CAST(unixepoch(date(month, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER) AS month_start,
    CAST(unixepoch(date(month, 'unixepoch', 'start of month', '+1 month')) AS UNIX_EPOCH_INTEGER) AS month_end
  FROM allocation
  UNION
  SELECT DISTINCT 
    CAST(unixepoch(date(t.date, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER) AS month_start,
    CAST(unixepoch(date(t.date, 'unixepoch', 'start of month', '+1 month')) AS UNIX_EPOCH_INTEGER) AS month_end
  FROM "transaction" AS t
),
income AS (
    SELECT
        t.date,
        tc.inflow AS net_inflow
    FROM transaction_category AS tc
    JOIN "transaction" AS t ON tc."transaction" = t.id
    WHERE tc.income
)
SELECT
   month_start AS month,
   (
     COALESCE((
      SELECT sum(total_inflow - total_outflow) 
      FROM "transaction"
      WHERE  date < month_end
     ), 0)
     - COALESCE((
      SELECT sum(
        CASE
          WHEN available > 0 THEN available
          ELSE 0
        END
      ) 
      FROM budget_month_categories 
      WHERE month = month_start
     ), 0)
   ) AS ready_to_assign,
   COALESCE((
    SELECT sum(net_inflow)
    FROM income 
    WHERE date >= month_start AND date < month_end
   ), 0) AS income,
   COALESCE((
    SELECT sum(allocated) 
    FROM budget_month_categories 
    WHERE month = month_start
   ), 0) AS allocated,
   COALESCE((
    SELECT sum(spent) 
    FROM budget_month_categories 
    WHERE month = month_start
   ), 0) AS spent,
   COALESCE((
    SELECT sum(available) 
    FROM budget_month_categories 
    WHERE month = month_start
   ), 0) AS available,
   (
    COALESCE((
      SELECT sum(total_inflow + total_outflow) 
      FROM "transaction"
      WHERE date >= month_start AND date < month_end
      ), 0) -
    COALESCE((
      SELECT sum(inflow + outflow) 
      FROM transaction_category AS tc
      JOIN  "transaction" AS t ON tc."transaction" = t.id
      WHERE t.date >= month_start AND t.date < month_end
      ), 0)
   ) AS uncategorized
FROM months;

-- +goose down

DROP VIEW budget_month_categories;
DROP VIEW budget_month_summary;
