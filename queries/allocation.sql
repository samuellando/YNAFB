-- name: CreateAllocation :one
INSERT INTO
  allocation (budget_id, category_id, month, amount)
SELECT
    b.id, ?, ?, ?
    FROM budget AS b
    WHERE b.id = @budget_id AND b.login_id = @login_id
RETURNING
  *;

-- name: UpdateAllocation :execrows
UPDATE allocation
SET
  amount = ?
WHERE
  allocation.category_id = @category_id
  AND allocation.month = ?
  AND allocation.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id
      AND b.login_id = @login_id
  );
