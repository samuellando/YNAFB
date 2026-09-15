-- name: CreateExpenseShare :one
INSERT INTO
  expense_share (default_name)
VALUES
  (?)
RETURNING
  *;

-- name: JoinExpenseShare :one 
INSERT INTO
  budget_expense_share (budget_id, expense_share_id, name)
SELECT
  b.id,
  @expense_share_id,
  (
    SELECT
      default_name
    FROM
      expense_share
    WHERE
      id = @expense_share_id
  ) AS name
FROM
  budget AS b
WHERE
  b.login_id = @login_id
  AND b.id = @budget_id
RETURNING
  *;

-- name: ListBudgetExpenseShares :many
SELECT
  e.id,
  e.default_name,
  bes.name
FROM
  expense_share AS e
  JOIN budget_expense_share AS bes ON e.id = bes.expense_share_id
  JOIN budget as b ON b.id = bes.budget_id
WHERE
  b.id = @budget_id
  AND b.login_id = @login_id;

-- name: CreateExpenseShareCode :one
INSERT INTO
  expense_share_code (expense_share_id, code, created)
SELECT
  e.id,
  @code,
  unixepoch()
FROM
  expense_share AS e
  JOIN budget_expense_share AS bes ON e.id = bes.expense_share_id
  JOIN budget as b ON b.id = bes.budget_id
WHERE
  b.id = @budget_id
  AND b.login_id = @login_id
  AND e.id = @expense_share_id
RETURNING
  *;

-- name: GetExpenseShareByCode :one
SELECT
  sqlc.embed(e), sqlc.embed(esc)
FROM
  expense_share AS e
  JOIN expense_share_code AS esc ON esc.expense_share_id = e.id
  AND esc.code = @code;

-- name: LeaveExpenseShare :exec
DELETE FROM budget_expense_share
WHERE
  expense_share_id = @expense_share_id
  AND budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id
      AND b.login_id = @login_id
  );
