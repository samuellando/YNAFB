-- name: CreateCategory :one
INSERT INTO
  category (budget_id, name, category_group_id)
SELECT b.id, ?, ?
FROM budget AS b
WHERE b.id = @budget_id AND b.login_id = @login_id
RETURNING
  *;

-- name: UpdateCategory :one
UPDATE category
SET
  name = ?,
  category_group_id = ?
WHERE
  category.id = @id
  AND category.budget_id IN (
    SELECT b.id
    FROM budget AS b
    WHERE b.id = @budget_id AND b.login_id = @login_id
)
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM category
WHERE
  category.id = @id
  AND category.budget_id IN (
    SELECT b.id
    FROM budget AS b
    WHERE b.id = @budget_id AND b.login_id = @login_id
);

-- name: ListCategories :many
SELECT
  c.id,
  c.budget_id,
  c.name,
  g.id AS group_id,
  g.name AS group_name
FROM
  category AS c
JOIN budget AS b ON c.budget_id = b.id
LEFT JOIN category_group AS g ON c.category_group_id = g.id
WHERE
  b.login_id = @login_id AND c.budget_id = @budget_id
ORDER BY
  c.name;

-- name: GetCategory :one
SELECT
  c.id,
  c.budget_id,
  c.name,
  g.id AS group_id,
  g.name AS group_name
FROM
  category AS c
JOIN budget AS b ON c.budget_id = b.id
LEFT JOIN category_group AS g ON c.category_group_id = g.id
WHERE
  b.login_id = @login_id
  AND c.budget_id = @budget_id
  AND c.id = @id;
