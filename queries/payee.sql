-- name: CreatePayee :one
INSERT INTO
  payee (budget_id, name)
SELECT
  b.id, ?
FROM
  budget AS b
WHERE
  b.id = @budget_id AND b.login_id = @login_id
RETURNING
  *;

-- name: UpdatePayee :execrows
UPDATE payee
SET
  name = ?
WHERE
  payee.id = @id
  AND payee.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id AND b.login_id = @login_id
  );

-- name: DeletePayee :exec
DELETE FROM payee
WHERE
  payee.id = @id
  AND payee.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id AND b.login_id = @login_id
  );

-- name: ListPayees :many
SELECT
  p.id,
  p.budget_id,
  p.name
FROM
  payee AS p
  JOIN budget AS b ON p.budget_id = b.id
WHERE
  b.login_id = @login_id AND p.budget_id = @budget_id
ORDER BY
  p.name;

-- name: GetPayeeByName :one
SELECT
  p.id,
  p.budget_id,
  p.name
FROM
  payee AS p
  JOIN budget AS b ON p.budget_id = b.id
WHERE
  b.login_id = @login_id
  AND p.budget_id = @budget_id
  AND p.name = ?;