-- name: GetAccountBalanceAsOf :one
SELECT
  (SELECT COALESCE(SUM(t.total_inflow - t.total_outflow), 0)
     FROM "transaction" AS t
    WHERE t.account = @account_id AND t.date <= @date)
  + (SELECT COALESCE(SUM(ts.outflow - ts.inflow), 0)
       FROM transaction_category AS ts
       JOIN "transaction" AS t ON t.id = ts."transaction"
      WHERE ts.other_account = @account_id AND t.date <= @date) AS balance;

-- name: ReconcileAccountTransactions :execrows
UPDATE "transaction"
SET reconciled = true
WHERE account = @account_id AND date <= @date AND reconciled = false;