-- +goose up

CREATE VIEW account_transactions AS
SELECT
  a.id as account_id,
  COALESCE(t.id, tcp.id)  as transaction_id,
  tc.id as transaction_category_id,
  COALESCE(t.date, tcp.date) as date,
  COALESCE(t.payee, tcp.payee) as payee,
  COALESCE(t.total_inflow, tc.outflow) as inflow,
  COALESCE(t.total_outflow, tc.inflow) as outflow,
  COALESCE(t.note, tcp.note, '') as note,
  (SELECT true FROM reconciliation as r WHERE r.account = a.id AND COALESCE(t.id, tcp.id) = r."transaction" LIMIT 1) as reconciled
FROM account AS a
LEFT JOIN "transaction" AS t ON t.account = a.id
LEFT JOIN "transaction_category" AS tc ON tc.other_account = a.id
LEFT JOIN "transaction" AS tcp ON tcp.id = tc."transaction";

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
JOIN account_transactions as at ON a.id = at.account_id
GROUP BY a.id, a.name;

-- +goose down

DROP VIEW account_transactions;
DROP VIEW account_balances;
