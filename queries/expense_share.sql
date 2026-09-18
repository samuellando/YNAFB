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

-- name: GetBudgetExpenseShare :one
SELECT
  bes.expense_share_id,
  bes.name,
  bes.display_name
FROM
  budget_expense_share AS bes
  JOIN budget as b ON b.id = bes.budget_id
WHERE
  bes.expense_share_id = @expense_share_id
  AND b.id = @budget_id
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
  src.account_id AS source_account_id,
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
  LEFT JOIN trx AS src ON src.id = est.trx_id
  AND src.budget_id = est.budget_id
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

-- name: ListExpenseShareBalances :many
-- Per-member net balances for an expense share, computed in SQL.
-- Positive = others owe this member; negative = this member owes.
-- Replicates the read-path default-split rule from groupExpenseShareTrx:
-- members without a stored split share the requested amount evenly in
-- budget-id order (leading members absorb leftover cents); the publisher,
-- while still a member, keeps total - requested. Stored splits of departed
-- or deleted budgets are ignored.
WITH gate AS (
  SELECT 1 AS ok
  FROM budget AS b
  JOIN budget_expense_share AS mine ON mine.budget_id = b.id AND mine.expense_share_id = @expense_share_id
  WHERE b.id = @budget_id AND b.login_id = @login_id
),
members AS (
  SELECT bes.budget_id, bes.display_name
  FROM gate
  JOIN budget_expense_share AS bes ON bes.expense_share_id = @expense_share_id
),
trx AS (
  SELECT
    est.id,
    est.budget_id AS publisher_budget_id,
    est.total_outflow,
    est.total_inflow,
    est.requested_outflow,
    est.requested_inflow
  FROM gate
  JOIN expense_share_trx AS est ON est.expense_share_id = @expense_share_id
),
stored AS (
  SELECT s.expense_share_trx_id AS trx_id, s.budget_id, s.split_outflow, s.split_inflow
  FROM expense_share_trx_split AS s
  JOIN members AS m ON m.budget_id = s.budget_id
  WHERE s.expense_share_id = @expense_share_id
),
owing AS (
  SELECT t.id AS trx_id, m.budget_id
  FROM trx AS t
  CROSS JOIN members AS m
  WHERE NOT EXISTS (
    SELECT 1 FROM stored AS s WHERE s.trx_id = t.id AND s.budget_id = m.budget_id
  )
),
sharing AS (
  SELECT o.trx_id, o.budget_id,
    ROW_NUMBER() OVER (PARTITION BY o.trx_id ORDER BY o.budget_id) AS rn
  FROM owing AS o
  JOIN trx AS t ON t.id = o.trx_id
  LEFT JOIN members AS pm ON pm.budget_id = t.publisher_budget_id
  WHERE t.publisher_budget_id IS NULL OR pm.budget_id IS NULL OR o.budget_id != t.publisher_budget_id
),
stats AS (
  SELECT
    t.id AS trx_id,
    CASE WHEN t.total_inflow = 0 THEN 1 ELSE 0 END AS outflow,
    CASE WHEN t.total_inflow = 0 THEN t.requested_outflow ELSE t.requested_inflow END AS requested,
    CASE WHEN t.total_inflow = 0 THEN t.total_outflow ELSE t.total_inflow END AS total,
    t.publisher_budget_id AS publisher_budget_id,
    CASE WHEN pm.budget_id IS NOT NULL THEN 1 ELSE 0 END AS pub_is_member
  FROM trx AS t
  LEFT JOIN members AS pm ON pm.budget_id = t.publisher_budget_id
),
counts AS (
  SELECT trx_id, COUNT(*) AS n FROM sharing GROUP BY trx_id
),
eff AS (
  SELECT
    t.trx_id,
    m.budget_id,
    COALESCE(
      CASE WHEN t.outflow = 1 THEN st.split_outflow ELSE st.split_inflow END,
      CASE WHEN t.pub_is_member = 1 AND m.budget_id = t.publisher_budget_id THEN t.total - t.requested END,
      CASE WHEN sh.rn IS NOT NULL
        THEN t.requested / c.n + CASE WHEN sh.rn <= t.requested % c.n THEN 1 ELSE 0 END
      END,
      0
    ) AS share
  FROM stats AS t
  CROSS JOIN members AS m
  LEFT JOIN stored AS st ON st.trx_id = t.trx_id AND st.budget_id = m.budget_id
  LEFT JOIN sharing AS sh ON sh.trx_id = t.trx_id AND sh.budget_id = m.budget_id
  LEFT JOIN counts AS c ON c.trx_id = t.trx_id
)
SELECT
  m.budget_id AS budget_id,
  m.display_name AS display_name,
  CAST(COALESCE(SUM(
    CASE WHEN st.outflow = 1
      THEN (CASE WHEN st.publisher_budget_id = m.budget_id THEN st.total ELSE 0 END) - sh.share
      ELSE sh.share - (CASE WHEN st.publisher_budget_id = m.budget_id THEN st.total ELSE 0 END)
    END
  ), 0) AS INTEGER) AS balance
FROM members AS m
LEFT JOIN eff AS sh ON sh.budget_id = m.budget_id
LEFT JOIN stats AS st ON st.trx_id = sh.trx_id
GROUP BY m.budget_id, m.display_name
ORDER BY balance DESC, m.budget_id;

-- name: GetExpenseShareTrxById :one
SELECT
  est.*
FROM
  budget AS b
  JOIN budget_expense_share AS mine ON mine.budget_id = b.id
  AND mine.expense_share_id = @expense_share_id
  JOIN expense_share_trx AS est ON est.expense_share_id = @expense_share_id
  AND est.id = @trx_id
WHERE
  b.id = @budget_id
  AND b.login_id = @login_id;

-- name: DeleteExpenseShareSplits :exec
DELETE FROM expense_share_trx_split
WHERE
  expense_share_trx_id = @trx_id;

-- name: CreateExpenseShareSplit :one
INSERT INTO
  expense_share_trx_split (
    expense_share_trx_id,
    expense_share_id,
    budget_id,
    split_outflow,
    split_inflow
  )
VALUES
  (@trx_id, @expense_share_id, @budget_id, @split_outflow, @split_inflow)
RETURNING
  *;

-- name: GetExpenseShareSplit :one
SELECT
  *
FROM
  expense_share_trx_split
WHERE
  expense_share_trx_id = @trx_id
  AND budget_id = @budget_id;

-- name: DeleteExpenseShareSplitLines :exec
DELETE FROM expense_share_trx_split_line
WHERE
  expense_share_trx_split_id = @split_id;

-- name: CreateExpenseShareSplitLine :one
INSERT INTO
  expense_share_trx_split_line (
    budget_id,
    expense_share_trx_split_id,
    dest_account_id,
    category_id,
    outflow,
    inflow
  )
VALUES
  (@budget_id, @split_id, @dest_account_id, @category_id, @outflow, @inflow)
RETURNING
  *;
