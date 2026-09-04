-- name: CreateAccount :one
INSERT INTO account (
  budget,
  name
) VALUES (
  ?,
  ?
)
RETURNING *;

-- name: UpdateAccount :execrows
UPDATE account
SET name = ?
WHERE id = ?;

-- name: DeleteAccount :exec
DELETE FROM account
WHERE id = ?;

-- name: GetAccountByName :one
SELECT id, budget, name
FROM account
WHERE budget = ? AND name = ?;

-- name: ListAccountsBalances :many
SELECT
   a.id,
   a.name,
   ab.balance,
   ab.reconciled_balance
FROM account AS a
JOIN account_balances as ab ON a.id = ab.id
WHERE a.budget = ?
GROUP BY a.id, a.name;

-- name: GetAccountBalances :one
SELECT
   a.id,
   a.name,
   ab.balance,
   ab.reconciled_balance
FROM account AS a
JOIN account_balances as ab ON a.id = ab.id
WHERE a.id = ?
GROUP BY a.id, a.name;

-- name: GetAccountBalanceAsOf :one
SELECT
  (SELECT COALESCE(SUM(t.total_inflow - t.total_outflow), 0)
     FROM "transaction" AS t
    WHERE t.account = @account_id AND t.date <= @date)
  + (SELECT COALESCE(SUM(ts.outflow - ts.inflow), 0)
       FROM transaction_category AS ts
       JOIN "transaction" AS t ON t.id = ts."transaction"
      WHERE ts.other_account = @account_id AND t.date <= @date) AS balance;

-- name: ListAccountTransactions :many
SELECT
  at.transaction_id as id,
  sqlc.embed(a),
  at.date,
  sqlc.embed(p),
  at.outflow as total_outflow,
  at.inflow as total_inflow,
  at.note,
  sa.id as source_account_id,
  sa.name as source_account_name,
  tc.id as transaction_category_id,
  oa.id as to_account_id,
  oa.name as to_account_name,
  c.id as category_id,
  c.name as category_name,
  COALESCE(tc.income, false),
  tc.outflow,
  tc.inflow,
  at.reconciled
FROM account_transactions AS at
LEFT JOIN "transaction_category" AS tc ON (at.transaction_id = tc."transaction" AND at.source_account IS NULL)
LEFT JOIN account AS a ON at.account_id = a.id
LEFT JOIN account AS sa ON at.source_account = sa.id
LEFT JOIN account AS oa ON tc.other_account = oa.id
LEFT JOIN category AS c ON tc.category = c.id
LEFT JOIN payee AS p ON at.payee = p.id
WHERE at.account_id = ?
ORDER BY at.date DESC, p.name ASC, at.transaction_id, tc.id;
