-- name: ReconcileAccountTransactions :execrows
INSERT INTO reconciliation (budget_id, account_id, trx_id, trx_line_id)
SELECT
  @budget_id,
  @id,
  CASE WHEN t.account_id = @id THEN t.id END AS trx_id,
  CASE WHEN tl.id IS NOT NULL AND t.account_id != @id THEN tl.id END AS trx_line_id
FROM trx AS t
JOIN budget AS b ON t.budget_id = b.id AND b.login_id = @login_id
LEFT JOIN trx_line AS tl ON tl.trx_id = t.id AND tl.dest_account_id = @id
WHERE t.budget_id = @budget_id
  AND t.date <= @date
  AND (t.account_id = @id OR tl.id IS NOT NULL)
  AND NOT EXISTS (
    SELECT 1 FROM reconciliation AS r
    WHERE r.account_id = @id AND (r.trx_id = t.id OR r.trx_line_id = tl.id)
  );
