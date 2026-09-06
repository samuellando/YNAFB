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

-- name: UpdateTrx :execrows
UPDATE trx
SET
  date = ?, account_id = ?, payee_id = ?, total_outflow = ?, total_inflow = ?, note = ?
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