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
SELECT
  a.*,
  c.name AS category_name,
  cg.id AS category_group_id,
  cg.name AS category_group_name
FROM allocation AS a
JOIN budget AS b ON b.id = a.budget_id
JOIN category AS c ON c.id = a.category_id
LEFT JOIN category_group AS cg ON cg.id = c.category_group_id
WHERE b.id = @budget_id AND b.login_id = @login_id
