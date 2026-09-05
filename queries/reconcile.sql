-- name: ReconcileAccountTransactions :execrows
INSERT INTO reconciliation (budget_id, account_id, trx_id, trx_line_id)
SELECT
  @budget_id,
  @account_id,
  CASE WHEN t.account_id = @account_id THEN t.id END AS trx_id,
  CASE WHEN tl.id IS NOT NULL AND t.account_id != @account_id THEN tl.id END AS trx_line_id
FROM trx AS t
LEFT JOIN trx_line AS tl ON tl.trx_id = t.id AND tl.dest_account_id = @account_id
WHERE t.budget_id = @budget_id
  AND t.date <= @date
  AND (t.account_id = @account_id OR tl.id IS NOT NULL)
  AND NOT EXISTS (
    SELECT 1 FROM reconciliation AS r
    WHERE r.account_id = @account_id AND (r.trx_id = t.id OR r.trx_line_id = tl.id)
  );