-- +goose up
DROP VIEW budget_month_categories;

DROP VIEW budget_month_summary;

CREATE VIEW budget_month_categories AS
WITH
  one AS (
    SELECT
      c.budget_id AS budget_id,
      m.month AS month,
      c.id AS category_id,
      CAST(COALESCE(a.allocated, 0) AS INTEGER) AS allocated,
      CAST(
        COALESCE(s.spent, 0) + COALESCE(es.spent, 0) AS INTEGER
      ) AS spent
    FROM
      category AS c
      JOIN (
        SELECT DISTINCT
          a.budget_id AS budget_id,
          CAST(
            unixepoch(date(a.month, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
          ) AS month
        FROM
          allocation AS a
        UNION
        SELECT DISTINCT
          t.budget_id AS budget_id,
          CAST(
            unixepoch(date(t.date, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
          ) AS month
        FROM
          trx AS t
          JOIN trx_line AS tc ON tc.trx_id = t.id
        WHERE
          tc.category_id IS NOT NULL
        UNION
        SELECT DISTINCT
          esl.budget_id AS budget_id,
          CAST(
            unixepoch(date(est.date, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
          ) AS month
        FROM
          expense_share_trx AS est
          JOIN expense_share_trx_split AS ess ON est.id = ess.expense_share_trx_id
          JOIN expense_share_trx_split_line AS esl ON ess.id = esl.expense_share_trx_split_id
        WHERE
          esl.category_id IS NOT NULL
      ) AS m ON m.budget_id = c.budget_id
      LEFT JOIN (
        SELECT
          a.category_id,
          CAST(
            unixepoch(date(a.month, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
          ) AS month,
          SUM(a.amount) AS allocated
        FROM
          allocation AS a
        GROUP BY
          a.category_id,
          month
      ) AS a ON a.category_id = c.id
      AND a.month = m.month
      LEFT JOIN (
        SELECT
          ts.category_id,
          CAST(
            unixepoch(date(t.date, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
          ) AS month,
          SUM(ts.outflow - ts.inflow) AS spent
        FROM
          trx_line AS ts
          JOIN trx AS t ON t.id = ts.trx_id
        WHERE
          ts.category_id IS NOT NULL
        GROUP BY
          ts.category_id,
          month
      ) AS s ON s.category_id = c.id
      AND s.month = m.month
      LEFT JOIN (
        SELECT
          esl.category_id,
          CAST(
            unixepoch(date(est.date, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
          ) AS month,
          SUM(esl.outflow - esl.inflow) AS spent
        FROM
          expense_share_trx AS est
          JOIN expense_share_trx_split AS ess ON est.id = ess.expense_share_trx_id
          JOIN expense_share_trx_split_line AS esl ON ess.id = esl.expense_share_trx_split_id
        WHERE
          esl.category_id IS NOT NULL
        GROUP BY
          esl.category_id,
          month
      ) AS es ON es.category_id = c.id
      AND es.month = m.month
  )
SELECT
  budget_id,
  month,
  category_id,
  allocated,
  spent,
  CAST(
    (
      CASE
        WHEN month < unixepoch() THEN SUM(MAX(allocated - spent, 0)) OVER (
          PARTITION by
            category_id
          ORDER BY
            month
        ) - MAX(allocated - spent, 0) + allocated - spent
        ELSE allocated - spent
      END
    ) AS INTEGER
  ) AS available
FROM
  one
GROUP BY
  month,
  category_id,
  budget_id
ORDER BY
  month,
  category_id;

CREATE VIEW budget_month_summary AS
WITH
  months AS (
    SELECT DISTINCT
      a.budget_id AS budget_id,
      CAST(
        unixepoch(date(a.month, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
      ) AS month_start,
      CAST(
        unixepoch(
          date(
            a.month,
            'unixepoch',
            'start of month',
            '+1 month'
          )
        ) AS UNIX_EPOCH_INTEGER
      ) AS month_end
    FROM
      allocation AS a
    UNION
    SELECT DISTINCT
      t.budget_id AS budget_id,
      CAST(
        unixepoch(date(t.date, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
      ) AS month_start,
      CAST(
        unixepoch(
          date(t.date, 'unixepoch', 'start of month', '+1 month')
        ) AS UNIX_EPOCH_INTEGER
      ) AS month_end
    FROM
      trx AS t
    UNION
    SELECT DISTINCT
      ess.budget_id AS budget_id,
      CAST(
        unixepoch(date(est.date, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
      ) AS month_start,
      CAST(
        unixepoch(
          date(
            est.date,
            'unixepoch',
            'start of month',
            '+1 month'
          )
        ) AS UNIX_EPOCH_INTEGER
      ) AS month_end
    FROM
      expense_share_trx AS est
      JOIN expense_share_trx_split AS ess ON est.id = ess.expense_share_trx_id
  ),
  income AS (
    SELECT
      t.budget_id AS budget_id,
      t.date,
      tc.inflow AS net_inflow
    FROM
      trx_line AS tc
      JOIN trx AS t ON tc.trx_id = t.id
    WHERE
      tc.income
  )
SELECT
  months.budget_id AS budget_id,
  month_start AS month,
  CAST(
    (
      COALESCE(
        (
          SELECT
            sum(total_inflow - total_outflow)
          FROM
            trx
          WHERE
            budget_id = months.budget_id
            AND date < month_end
        ),
        0
      ) - COALESCE(
        (
          SELECT
            sum(
              CASE
                WHEN available > 0 THEN available
                ELSE 0
              END
            )
          FROM
            budget_month_categories
          WHERE
            budget_id = months.budget_id
            AND month = month_start
        ),
        0
      )
    ) AS INTEGER
  ) AS ready_to_assign,
  CAST(
    COALESCE(
      (
        SELECT
          sum(net_inflow)
        FROM
          income
        WHERE
          budget_id = months.budget_id
          AND date >= month_start
          AND date < month_end
      ),
      0
    ) AS INTEGER
  ) AS income,
  CAST(
    COALESCE(
      (
        SELECT
          sum(allocated)
        FROM
          budget_month_categories
        WHERE
          budget_id = months.budget_id
          AND month = month_start
      ),
      0
    ) AS INTEGER
  ) AS allocated,
  CAST(
    COALESCE(
      (
        SELECT
          sum(spent)
        FROM
          budget_month_categories
        WHERE
          budget_id = months.budget_id
          AND month = month_start
      ),
      0
    ) AS INTEGER
  ) AS spent,
  CAST(
    COALESCE(
      (
        SELECT
          sum(available)
        FROM
          budget_month_categories
        WHERE
          budget_id = months.budget_id
          AND month = month_start
      ),
      0
    ) AS INTEGER
  ) AS available,
  CAST(
    (
      COALESCE(
        (
          SELECT
            sum(total_inflow + total_outflow)
          FROM
            trx
          WHERE
            budget_id = months.budget_id
            AND date >= month_start
            AND date < month_end
        ),
        0
      ) - COALESCE(
        (
          SELECT
            sum(inflow + outflow)
          FROM
            trx_line AS tc
            JOIN trx AS t ON tc.trx_id = t.id
          WHERE
            t.budget_id = months.budget_id
            AND t.date >= month_start
            AND t.date < month_end
        ),
        0
      )
    ) + (
      COALESCE(
        (
          SELECT
            sum(total_inflow + total_outflow)
          FROM
            expense_share_trx_split AS ess
            JOIN expense_share_trx AS est ON ess.expense_share_trx_id = est.id
          WHERE
            ess.budget_id = months.budget_id
            AND est.date >= month_start
            AND est.date < month_end
        ),
        0
      ) - COALESCE(
        (
          SELECT
            sum(inflow + outflow)
          FROM
            expense_share_trx AS est
            JOIN expense_share_trx_split AS ess ON est.id = ess.expense_share_trx_id
            JOIN expense_share_trx_split_line AS esl ON ess.id = esl.expense_share_trx_split_id
          WHERE
            ess.budget_id = months.budget_id
            AND est.date >= month_start
            AND est.date < month_end
        ),
        0
      )
    ) AS INTEGER
  ) AS uncategorized
FROM
  months;

-- +goose down
DROP VIEW budget_month_categories;

DROP VIEW budget_month_summary;

CREATE VIEW budget_month_categories AS
WITH
  one AS (
    SELECT
      c.budget_id AS budget_id,
      m.month AS month,
      c.id AS category_id,
      CAST(COALESCE(a.allocated, 0) AS INTEGER) AS allocated,
      CAST(COALESCE(s.spent, 0) AS INTEGER) AS spent
    FROM
      category AS c
      JOIN (
        SELECT DISTINCT
          a.budget_id AS budget_id,
          CAST(
            unixepoch(date(a.month, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
          ) AS month
        FROM
          allocation AS a
        UNION
        SELECT DISTINCT
          t.budget_id AS budget_id,
          CAST(
            unixepoch(date(t.date, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
          ) AS month
        FROM
          trx AS t
          JOIN trx_line AS tc ON tc.trx_id = t.id
        WHERE
          tc.category_id IS NOT NULL
      ) AS m ON m.budget_id = c.budget_id
      LEFT JOIN (
        SELECT
          a.category_id,
          CAST(
            unixepoch(date(a.month, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
          ) AS month,
          SUM(a.amount) AS allocated
        FROM
          allocation AS a
        GROUP BY
          a.category_id,
          month
      ) AS a ON a.category_id = c.id
      AND a.month = m.month
      LEFT JOIN (
        SELECT
          ts.category_id,
          CAST(
            unixepoch(date(t.date, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
          ) AS month,
          SUM(ts.outflow - ts.inflow) AS spent
        FROM
          trx_line AS ts
          JOIN trx AS t ON t.id = ts.trx_id
        WHERE
          ts.category_id IS NOT NULL
        GROUP BY
          ts.category_id,
          month
      ) AS s ON s.category_id = c.id
      AND s.month = m.month
  )
SELECT
  budget_id,
  month,
  category_id,
  allocated,
  spent,
  CAST(
    (
      CASE
        WHEN month < unixepoch() THEN SUM(MAX(allocated - spent, 0)) OVER (
          PARTITION by
            category_id
          ORDER BY
            month
        ) - MAX(allocated - spent, 0) + allocated - spent
        ELSE allocated - spent
      END
    ) AS INTEGER
  ) AS available
FROM
  one
GROUP BY
  month,
  category_id,
  budget_id
ORDER BY
  month,
  category_id;

CREATE VIEW budget_month_summary AS
WITH
  months AS (
    SELECT DISTINCT
      a.budget_id AS budget_id,
      CAST(
        unixepoch(date(a.month, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
      ) AS month_start,
      CAST(
        unixepoch(
          date(
            a.month,
            'unixepoch',
            'start of month',
            '+1 month'
          )
        ) AS UNIX_EPOCH_INTEGER
      ) AS month_end
    FROM
      allocation AS a
    UNION
    SELECT DISTINCT
      t.budget_id AS budget_id,
      CAST(
        unixepoch(date(t.date, 'unixepoch', 'start of month')) AS UNIX_EPOCH_INTEGER
      ) AS month_start,
      CAST(
        unixepoch(
          date(t.date, 'unixepoch', 'start of month', '+1 month')
        ) AS UNIX_EPOCH_INTEGER
      ) AS month_end
    FROM
      trx AS t
  ),
  income AS (
    SELECT
      t.budget_id AS budget_id,
      t.date,
      tc.inflow AS net_inflow
    FROM
      trx_line AS tc
      JOIN trx AS t ON tc.trx_id = t.id
    WHERE
      tc.income
  )
SELECT
  months.budget_id AS budget_id,
  month_start AS month,
  CAST(
    (
      COALESCE(
        (
          SELECT
            sum(total_inflow - total_outflow)
          FROM
            trx
          WHERE
            budget_id = months.budget_id
            AND date < month_end
        ),
        0
      ) - COALESCE(
        (
          SELECT
            sum(
              CASE
                WHEN available > 0 THEN available
                ELSE 0
              END
            )
          FROM
            budget_month_categories
          WHERE
            budget_id = months.budget_id
            AND month = month_start
        ),
        0
      )
    ) AS INTEGER
  ) AS ready_to_assign,
  CAST(
    COALESCE(
      (
        SELECT
          sum(net_inflow)
        FROM
          income
        WHERE
          budget_id = months.budget_id
          AND date >= month_start
          AND date < month_end
      ),
      0
    ) AS INTEGER
  ) AS income,
  CAST(
    COALESCE(
      (
        SELECT
          sum(allocated)
        FROM
          budget_month_categories
        WHERE
          budget_id = months.budget_id
          AND month = month_start
      ),
      0
    ) AS INTEGER
  ) AS allocated,
  CAST(
    COALESCE(
      (
        SELECT
          sum(spent)
        FROM
          budget_month_categories
        WHERE
          budget_id = months.budget_id
          AND month = month_start
      ),
      0
    ) AS INTEGER
  ) AS spent,
  CAST(
    COALESCE(
      (
        SELECT
          sum(available)
        FROM
          budget_month_categories
        WHERE
          budget_id = months.budget_id
          AND month = month_start
      ),
      0
    ) AS INTEGER
  ) AS available,
  CAST(
    (
      COALESCE(
        (
          SELECT
            sum(total_inflow + total_outflow)
          FROM
            trx
          WHERE
            budget_id = months.budget_id
            AND date >= month_start
            AND date < month_end
        ),
        0
      ) - COALESCE(
        (
          SELECT
            sum(inflow + outflow)
          FROM
            trx_line AS tc
            JOIN trx AS t ON tc.trx_id = t.id
          WHERE
            t.budget_id = months.budget_id
            AND t.date >= month_start
            AND t.date < month_end
        ),
        0
      )
    ) AS INTEGER
  ) AS uncategorized
FROM
  months;
