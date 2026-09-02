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

-- name: ListAccounts :many
SELECT id, budget, name
FROM account
WHERE budget = ?
ORDER BY name;

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
