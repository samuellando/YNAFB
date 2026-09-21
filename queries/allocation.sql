-- name: SetAllocation :one
INSERT OR REPLACE INTO
  allocation (budget_id, category_id, month, amount)
SELECT
    b.id, ?, ?, ?
    FROM budget AS b
    WHERE b.id = @budget_id AND b.login_id = @login_id
RETURNING
  *;

-- name: ListAllocations :many
SELECT a.*
FROM allocation AS a
JOIN budget AS b ON b.id = a.budget_id
WHERE b.id = @budget_id AND b.login_id = @login_id
