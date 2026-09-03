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
