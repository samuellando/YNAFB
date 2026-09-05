-- +goose up
CREATE VIEW account_trx AS
SELECT
  a.budget_id as budget_id,
  a.id as account_id,
  t.id as trx_id,
  CAST(NULL AS INTEGER) as source_account_id,
  t.date as date,
  t.payee_id as payee_id,
  t.total_inflow as inflow,
  t.total_outflow as outflow,
  COALESCE(t.note, '') as note,
  EXISTS (
    SELECT
      true
    FROM
      reconciliation as r
    WHERE
      r.account_id = a.id
      AND t.id = r.trx_id
    LIMIT
      1
  ) as reconciled
FROM
  account AS a
  JOIN trx AS t ON t.account_id = a.id
UNION ALL
SELECT
  a.budget_id as budget_id,
  a.id as account_id,
  tcp.id as trx_id,
  tcp.account_id as source_account_id,
  tcp.date as date,
  tcp.payee_id as payee_id,
  tc.outflow as inflow,
  tc.inflow as outflow,
  COALESCE(tcp.note, '') as note,
  EXISTS (
    SELECT
      true
    FROM
      reconciliation as r
    WHERE
      r.account_id = a.id
      AND tc.id = r.trx_line_id
    LIMIT
      1
  ) as reconciled
FROM
  account AS a
  JOIN trx_line AS tc ON tc.dest_account_id = a.id
  JOIN trx AS tcp ON tcp.id = tc.trx_id;

CREATE VIEW account_balances AS
SELECT
  a.budget_id,
  a.id AS account_id,
  a.name,
  CAST(
    COALESCE(SUM(at.inflow - at.outflow), 0) AS INTEGER
  ) AS balance,
  CAST(
    COALESCE(
      SUM(
        CASE
          WHEN at.reconciled THEN at.inflow - at.outflow
          ELSE 0
        END
      ),
      0
    ) AS INTEGER
  ) AS reconciled_balance
FROM
  account AS a
  LEFT JOIN account_trx as at ON a.id = at.account_id
GROUP BY
  a.budget_id,
  a.id,
  a.name;

-- +goose down
DROP VIEW account_trx;

DROP VIEW account_balances;
