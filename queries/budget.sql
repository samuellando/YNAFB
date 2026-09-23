-- name: CreateBudget :one
INSERT INTO budget (
  login_id,
  name
) VALUES (
  ?,
  ?
)
RETURNING *;

-- name: GetBudget :one
SELECT * 
FROM budget
WHERE id = @id AND login_id = @login_id;

-- name: UpdateBudget :one
UPDATE budget
SET
  name = ?
WHERE
  budget.login_id = @login_id AND budget.id = @id
RETURNING *;

-- name: DeleteBudget :exec
DELETE FROM budget
WHERE
  budget.login_id = @login_id AND budget.id = @id;

-- name: ListBudgets :many
SELECT
  budget.id,
  budget.login_id,
  budget.name
FROM
  budget
WHERE
  budget.login_id = @login_id
ORDER BY
  budget.id;
