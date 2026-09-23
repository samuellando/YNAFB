-- name: CreateTrx :one
INSERT INTO
  trx (budget_id, account_id, payee_id, date, total_outflow, total_inflow, note)
SELECT
  b.id, ?, ?, ?, ?, ?, ?
FROM
  budget AS b
WHERE
  b.id = @budget_id AND b.login_id = @login_id
RETURNING
  *;

-- name: UpdateTrx :one
UPDATE trx
SET
  date = ?, payee_id = ?, total_outflow = ?, total_inflow = ?, note = ?
WHERE
  trx.id = @id
  AND trx.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id AND b.login_id = @login_id
  )
RETURNING *;

-- name: DeleteTrx :exec
DELETE FROM trx
WHERE
  trx.id = @id
  AND trx.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id AND b.login_id = @login_id
  );

-- name: GetTrxAndLines :many
SELECT
    sqlc.embed(t),
    a.name AS account_name,
    p.name AS payee_name,
    tl.id AS line_id,
    tl.income AS line_income,
    tl.inflow AS line_inflow,
    tl.outflow AS line_outflow,
    c.id AS category_id,
    c.name AS category_name,
    cg.id AS category_group_id,
    cg.name AS category_group_name,
    da.id AS dest_account_id,
    da.name AS dest_account_name,
    CAST((r.account_id IS NOT NULL) AS bool) AS reconciled
FROM trx AS t
JOIN budget AS b ON t.budget_id = b.id
JOIN account AS a ON t.account_id = a.id
JOIN payee AS p ON p.id = t.payee_id
LEFT JOIN trx_line AS tl ON t.id = tl.trx_id
LEFT JOIN category AS c ON tl.category_id = c.id
LEFT JOIN category_group AS cg ON c.category_group_id = cg.id
LEFT JOIN account AS da ON tl.dest_account_id = da.id
LEFT JOIN reconciliation AS r ON r.account_id = a.id AND r.trx_id = t.id
WHERE b.login_id = @login_id AND t.budget_id = @budget_id AND t.id = @id AND a.id = @account_id
ORDER BY t.date DESC, t.id DESC;

-- name: ListTrxs :many
SELECT
  t.id,
  t.budget_id,
  t.date,
  a.name AS account_name,
  p.name AS payee_name,
  t.total_outflow,
  t.total_inflow,
  t.note
FROM trx AS t
JOIN account AS a ON a.id = t.account_id
JOIN payee AS p ON p.id = t.payee_id
JOIN budget AS b ON t.budget_id = b.id
WHERE b.login_id = @login_id AND t.budget_id = @budget_id
ORDER BY t.date DESC, t.id DESC;

-- name: ListTrxsAndLines :many
SELECT
    sqlc.embed(t),
    a.name AS account_name,
    p.name AS payee_name,
    tl.id AS line_id,
    tl.income AS line_income,
    tl.inflow AS line_inflow,
    tl.outflow AS line_outflow,
    c.id AS category_id,
    c.name AS category_name,
    cg.id AS category_group_id,
    cg.name AS category_group_name,
    da.id AS dest_account_id,
    da.name AS dest_account_name,
    CAST((r.account_id IS NOT NULL) AS bool) AS reconciled
FROM trx AS t
JOIN budget AS b ON t.budget_id = b.id
JOIN account AS a ON t.account_id = a.id
JOIN payee AS p ON p.id = t.payee_id
LEFT JOIN trx_line AS tl ON t.id = tl.trx_id
LEFT JOIN category AS c ON tl.category_id = c.id
LEFT JOIN category_group AS cg ON c.category_group_id = cg.id
LEFT JOIN account AS da ON tl.dest_account_id = da.id
LEFT JOIN reconciliation AS r ON r.account_id = a.id AND r.trx_id = t.id
WHERE b.login_id = @login_id AND t.budget_id = @budget_id
ORDER BY t.date DESC, t.id DESC;
