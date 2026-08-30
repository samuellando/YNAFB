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
INSERT OR IGNORE INTO reconciliation (account, "transaction")
SELECT @account_id, t.id
  FROM "transaction" AS t
 WHERE t.date <= @date
   AND (t.account = @account_id
        OR EXISTS (
          SELECT 1 FROM transaction_category AS ts
          WHERE ts."transaction" = t.id AND ts.other_account = @account_id
        ));