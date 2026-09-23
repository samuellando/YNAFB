-- name: CreateAccount :one
INSERT INTO
  account (budget_id, name)
SELECT
  b.id,
  @name
FROM
  budget AS b
WHERE
  b.login_id = @login_id
  AND b.id = @budget_id
RETURNING
  id, budget_id, name,
  (
    SELECT
      login_id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id
  ),
  (
    SELECT
      name AS budget_name
    FROM
      budget AS b
    WHERE
      b.id = @budget_id
  );

-- name: UpdateAccount :one
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
  )
RETURNING
  *;

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

-- name: GetAccount :one
SELECT
  b.login_id AS login_id,
  b.name AS budget_name,
  account.*
FROM
  account
  JOIN budget as b ON account.budget_id = b.id
WHERE
  b.login_id = @login_id
  AND account.budget_id = @budget_id
  AND account.id = @id;

-- name: ListAccounts :many
SELECT
  b.login_id AS login_id,
  b.name AS budget_name,
  account.*
FROM
  account
  JOIN budget as b ON account.budget_id = b.id
WHERE
  b.login_id = @login_id
  AND account.budget_id = @budget_id
ORDER BY
  account.name;
