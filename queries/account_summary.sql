-- name: GetAccountBalances :one
SELECT
  (SELECT COALESCE(SUM(t.total_inflow - t.total_outflow), 0) FROM "transaction" AS t WHERE t.account = @account_id)
  + (SELECT COALESCE(SUM(ts.outflow - ts.inflow), 0) FROM transaction_category AS ts WHERE ts.other_account = @account_id) AS balance,
  (SELECT COALESCE(SUM(t.total_inflow - t.total_outflow), 0)
     FROM "transaction" AS t
     JOIN reconciliation AS r ON r."transaction" = t.id AND r.account = @account_id
    WHERE t.account = @account_id)
  + (SELECT COALESCE(SUM(ts.outflow - ts.inflow), 0)
     FROM transaction_category AS ts
     JOIN "transaction" AS t ON t.id = ts."transaction"
     JOIN reconciliation AS r ON r."transaction" = t.id AND r.account = @account_id
    WHERE ts.other_account = @account_id) AS reconciled_balance;

-- name: ListAccountBalances :many
SELECT
  a.id,
  a.name,
  (SELECT COALESCE(SUM(t.total_inflow - t.total_outflow), 0) FROM "transaction" AS t WHERE t.account = a.id)
  + (SELECT COALESCE(SUM(ts.outflow - ts.inflow), 0) FROM transaction_category AS ts WHERE ts.other_account = a.id) AS balance,
  (SELECT COALESCE(SUM(t.total_inflow - t.total_outflow), 0)
     FROM "transaction" AS t
     JOIN reconciliation AS r ON r."transaction" = t.id AND r.account = a.id
    WHERE t.account = a.id)
  + (SELECT COALESCE(SUM(ts.outflow - ts.inflow), 0)
     FROM transaction_category AS ts
     JOIN "transaction" AS t ON t.id = ts."transaction"
     JOIN reconciliation AS r ON r."transaction" = t.id AND r.account = a.id
    WHERE ts.other_account = a.id) AS reconciled_balance
FROM account AS a
WHERE a.budget = ?
ORDER BY a.name;

-- name: ListBudgetBalances :many
SELECT
  b.id,
  b.name,
  (SELECT COALESCE(SUM(t.total_inflow - t.total_outflow), 0)
   FROM "transaction" AS t
   JOIN account AS a ON a.id = t.account
   WHERE a.budget = b.id)
  + (SELECT COALESCE(SUM(ts.outflow - ts.inflow), 0)
     FROM transaction_category AS ts
     JOIN account AS a ON a.id = ts.other_account
     WHERE a.budget = b.id) AS balance,
  (SELECT COALESCE(SUM(t.total_inflow - t.total_outflow), 0)
   FROM "transaction" AS t
   JOIN account AS a ON a.id = t.account
   JOIN reconciliation AS r ON r."transaction" = t.id AND r.account = a.id
   WHERE a.budget = b.id)
  + (SELECT COALESCE(SUM(ts.outflow - ts.inflow), 0)
     FROM transaction_category AS ts
     JOIN account AS a ON a.id = ts.other_account
     JOIN "transaction" AS t ON t.id = ts."transaction"
     JOIN reconciliation AS r ON r."transaction" = t.id AND r.account = a.id
     WHERE a.budget = b.id) AS reconciled_balance
FROM budget AS b
ORDER BY b.name;