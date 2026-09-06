-- name: CreateAccount :one
INSERT INTO
  account (budget_id, name)
SELECT
  b.id,
  ?
FROM
  budget AS b
WHERE
  b.login_id = @login_id
  AND b.id = @budget_id
RETURNING
  *;

-- name: UpdateAccount :execrows
UPDATE account
SET
  name = ?
WHERE
  account.id = @id
  AND account.budget_id IN (
    SELECT
      b.id
    FROM
      budget as b
    WHERE
      b.login_id = @login_id
      AND b.id = @budget_id
  );

-- name: DeleteAccount :exec
DELETE FROM account
WHERE
  account.id = @id
  AND account.budget_id IN (
    SELECT
      b.id
    FROM
      budget as b
    WHERE
      b.login_id = @login_id
      AND b.id = @budget_id
  );

-- name: GetAccountByName :one
SELECT
  account.id,
  account.budget_id,
  account.name
FROM
  account
  JOIN budget as b ON account.budget_id = b.id
WHERE
  b.login_id = @login_id
  AND account.budget_id = @budget_id
  AND account.name = ?;

-- name: ListAccountsBalances :many
SELECT
  a.id,
  a.budget_id,
  a.name,
  ab.balance,
  ab.reconciled_balance
FROM
  account AS a
  JOIN account_balances as ab ON a.id = ab.account_id
  JOIN budget as b ON a.budget_id = b.id
WHERE
  b.login_id = @login_id
  AND a.budget_id = @budget_id
ORDER BY
  a.name;

-- name: GetAccountBalances :one
SELECT
  a.id,
  a.budget_id,
  a.name,
  ab.balance,
  ab.reconciled_balance
FROM
  account AS a
  JOIN account_balances as ab ON a.id = ab.account_id
  JOIN budget as b ON a.budget_id = b.id
WHERE
  b.login_id = @login_id
  AND a.budget_id = @budget_id
  AND a.id = @id;

-- name: GetAccountBalanceAsOf :one
SELECT
  (
    SELECT
      COALESCE(SUM(t.total_inflow - t.total_outflow), 0)
    FROM
      trx AS t
      JOIN budget AS b ON t.budget_id = b.id
    WHERE
      b.login_id = @login_id
      AND t.budget_id = @budget_id
      AND t.account_id = @id
      AND t.date <= @date
  ) + (
    SELECT
      COALESCE(SUM(ts.outflow - ts.inflow), 0)
    FROM
      trx_line AS ts
      JOIN trx AS t ON t.id = ts.trx_id
      JOIN budget AS b ON t.budget_id = b.id
    WHERE
      b.login_id = @login_id
      AND ts.budget_id = @budget_id
      AND ts.dest_account_id = @id
      AND t.date <= @date
  ) AS balance;

-- name: ListAccountTransactions :many
SELECT
  at.budget_id,
  at.account_id,
  at.trx_id,
  at.source_account_id,
  at.date,
  at.payee_id,
  at.outflow,
  at.inflow,
  at.note,
  at.reconciled,
  tc.id AS trx_line_id,
  tc.dest_account_id,
  tc.category_id,
  COALESCE(tc.income, false),
  tc.outflow AS line_outflow,
  tc.inflow AS line_inflow,
  a.name AS account_name,
  sa.name AS source_account_name,
  oa.name AS dest_account_name,
  c.name AS category_name,
  p.name AS payee_name
FROM
  account_trx AS at
  JOIN budget AS b ON at.budget_id = b.id
  LEFT JOIN trx_line AS tc ON (
    at.trx_id = tc.trx_id
    AND at.source_account_id IS NULL
  )
  LEFT JOIN account AS a ON at.account_id = a.id
  LEFT JOIN account AS sa ON at.source_account_id = sa.id
  LEFT JOIN account AS oa ON tc.dest_account_id = oa.id
  LEFT JOIN category AS c ON tc.category_id = c.id
  LEFT JOIN payee AS p ON at.payee_id = p.id
WHERE
  b.login_id = @login_id
  AND at.budget_id = @budget_id
  AND at.account_id = @id
ORDER BY
  at.date DESC,
  p.name ASC,
  at.trx_id,
  tc.id;
