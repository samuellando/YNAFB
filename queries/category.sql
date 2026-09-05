-- name: CreateCategory :one
INSERT INTO category (
  budget_id,
  name,
  category_group_id
) VALUES (
  ?,
  ?,
  ?
)
RETURNING *;

-- name: UpdateCategory :execrows
UPDATE category
SET name = ?, category_group_id = ?
WHERE id = ? AND budget_id = ?;

-- name: DeleteCategory :exec
DELETE FROM category
WHERE id = ? AND budget_id = ?;

-- name: ListCategories :many
SELECT id, budget_id, name
FROM category
WHERE budget_id = ?
ORDER BY name;

-- name: GetCategoryByName :one
SELECT id, budget_id, name
FROM category
WHERE budget_id = ? AND name = ?;

-- name: CreateCategoryGroup :one
INSERT INTO category_group (
  budget_id,
  name
) VALUES (
  ?,
  ?
)
RETURNING id;

-- name: UpdateCategoryGroup :execrows
UPDATE category_group
SET name = ?
WHERE id = ? AND budget_id = ?;

-- name: GetOrCreateCategoryGroup :one
INSERT INTO category_group (
  budget_id,
  name
) VALUES (
  ?,
  ?
)
ON CONFLICT (budget_id, name) DO UPDATE SET name = excluded.name
RETURNING id;

-- name: DeleteCategoryGroup :exec
DELETE FROM category_group
WHERE id = ? AND budget_id = ?;

-- name: ListCategoryGroups :many
SELECT id, budget_id, name
FROM category_group
WHERE budget_id = ?
ORDER BY name;

-- name: GetCategoryGroupByName :one
SELECT id, budget_id, name
FROM category_group
WHERE budget_id = ? AND name = ?;