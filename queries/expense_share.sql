-- name: CreateExpenseShare :one
INSERT INTO
  expense_share (default_name)
VALUES
  (?)
RETURNING
  *;

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
  sqlc.embed(e),
  sqlc.embed(esc)
FROM
  expense_share AS e
  JOIN expense_share_code AS esc ON esc.expense_share_id = e.id
  AND esc.code = @code;

-- name: JoinExpenseShare :one 
INSERT INTO
  budget_expense_share (budget_id, expense_share_id, name, display_name)
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
  ) AS name,
  @display_name
FROM
  budget AS b
WHERE
  b.login_id = @login_id
  AND b.id = @budget_id
RETURNING
  *;

-- name: UpdateExpenseShare :one
UPDATE budget_expense_share
SET
  name = ?,
  display_name = ?
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
  )
RETURNING *;

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

-- name: ListBudgetExpenseShares :many
SELECT
  bes.expense_share_id,
  bes.name,
  bes.display_name
FROM
  budget_expense_share AS bes
  JOIN budget as b ON b.id = bes.budget_id
WHERE
  b.id = @budget_id
  AND b.login_id = @login_id;

-- name: PublishExpenseShareTrx :one
INSERT INTO
  expense_share_trx (
    expense_share_id,
    budget_id,
    trx_id,
    payee_name,
    date,
    total_outflow,
    total_inflow,
    requested_outflow,
    requested_inflow,
    note
  )
SELECT
  @expense_share_id,
  b.id,
  t.id,
  p.name,
  t.date,
  t.total_outflow,
  t.total_inflow,
  SUM(tl.outflow),
  SUM(tl.inflow),
  t.note
FROM
  budget AS b
  JOIN trx AS t ON t.budget_id = b.id AND t.id = @trx_id
  JOIN payee AS p ON p.budget_id = b.id AND p.id = t.payee_id
  JOIN budget_expense_share AS bes ON bes.budget_id = b.id
  AND bes.expense_share_id = @expense_share_id
  JOIN trx_line AS tl ON tl.budget_id = b.id
  AND tl.trx_id = t.id
  AND tl.expense_share_id = @expense_share_id
WHERE
  b.id = @budget_id
  AND b.login_id = @login_id
GROUP BY
  t.id
RETURNING
  *;

-- name: ListExpenseShareMembers :many
SELECT
  bes.budget_id,
  bes.display_name
FROM
  budget AS b
  JOIN budget_expense_share AS mine ON mine.budget_id = b.id
  AND mine.expense_share_id = @expense_share_id
  JOIN budget_expense_share AS bes ON bes.expense_share_id = @expense_share_id
WHERE
  b.id = @budget_id
  AND b.login_id = @login_id
ORDER BY
  bes.budget_id;

-- name: ListExpenseShareTrxDetails :many
SELECT
  est.id AS est_id,
  est.expense_share_id,
  est.budget_id AS publisher_budget_id,
  est.trx_id AS source_trx_id,
  est.payee_name,
  est.date,
  est.total_outflow,
  est.total_inflow,
  est.requested_outflow,
  est.requested_inflow,
  est.note,
  pub_bes.display_name AS publisher_display_name,
  ess_all.id AS split_id,
  ess_all.budget_id AS split_budget_id,
  split_bes.display_name AS split_display_name,
  ess_all.split_outflow,
  ess_all.split_inflow,
  esl.id AS line_id,
  esl.category_id,
  c.name AS category_name,
  esl.dest_account_id,
  oa.name AS dest_account_name,
  esl.outflow AS line_outflow,
  esl.inflow AS line_inflow
FROM
  budget AS b
  JOIN budget_expense_share AS mine ON mine.budget_id = b.id
  AND mine.expense_share_id = @expense_share_id
  JOIN expense_share_trx AS est ON est.expense_share_id = @expense_share_id
  LEFT JOIN budget_expense_share AS pub_bes ON pub_bes.budget_id = est.budget_id
  AND pub_bes.expense_share_id = @expense_share_id
  LEFT JOIN expense_share_trx_split AS ess_all ON ess_all.expense_share_trx_id = est.id
  LEFT JOIN budget_expense_share AS split_bes ON split_bes.budget_id = ess_all.budget_id
  AND split_bes.expense_share_id = @expense_share_id
  LEFT JOIN expense_share_trx_split AS ess_own ON ess_own.expense_share_trx_id = est.id
  AND ess_own.budget_id = @budget_id
  LEFT JOIN expense_share_trx_split_line AS esl ON esl.expense_share_trx_split_id = ess_own.id
  LEFT JOIN category AS c ON c.id = esl.category_id
  LEFT JOIN account AS oa ON oa.id = esl.dest_account_id
WHERE
  b.id = @budget_id
  AND b.login_id = @login_id
ORDER BY
  est.date DESC,
  est.id DESC,
  ess_all.budget_id,
  esl.id;
