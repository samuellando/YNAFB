-- +goose up

CREATE VIEW account_transactions AS
SELECT
  a.id as account_id,
  t.id  as transaction_id,
  CAST(NULL AS INTEGER) as source_account,
  t.date as date,
  t.payee as payee,
  t.total_inflow as inflow,
  t.total_outflow as outflow,
  COALESCE(t.note, '') as note,
  EXISTS(SELECT true FROM reconciliation as r WHERE r.account = a.id AND t.id = r."transaction" LIMIT 1) as reconciled
FROM account AS a
JOIN "transaction" AS t ON t.account = a.id
UNION ALL
SELECT
  a.id as account_id,
  tcp.id  as transaction_id,
  tcp.account as source_account,
  tcp.date as date,
  tcp.payee as payee,
  tc.outflow as inflow,
  tc.inflow as outflow,
  COALESCE(tcp.note, '') as note,
  EXISTS(SELECT true FROM reconciliation as r WHERE r.account = a.id AND tcp.id = r."transaction" LIMIT 1) as reconciled
FROM account AS a
JOIN "transaction_category" AS tc ON tc.other_account = a.id
JOIN "transaction" AS tcp ON tcp.id = tc."transaction";

CREATE VIEW account_balances AS
SELECT
   a.id,
   a.name,
   CAST(COALESCE(SUM(at.inflow - at.outflow), 0) AS INTEGER) AS balance,
   CAST(COALESCE(SUM(CASE 
        WHEN at.reconciled THEN at.inflow - at.outflow
        ELSE 0
   END), 0) AS INTEGER) AS reconciled_balance
FROM account AS a
LEFT JOIN account_transactions as at ON a.id = at.account_id
GROUP BY a.id, a.name;

-- +goose down

DROP VIEW account_transactions;
DROP VIEW account_balances;
