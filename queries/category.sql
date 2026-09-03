-- name: CreateCategory :one
INSERT INTO category (
  budget,
  name,
  category_group
) VALUES (
  ?,
  ?,
  ?
)
RETURNING *;

-- name: UpdateCategory :execrows
UPDATE category
SET name = ?, category_group = ?
WHERE id = ?;

-- name: DeleteCategory :exec
DELETE FROM category
WHERE id = ?;

-- name: ListCategories :many
SELECT id, budget, name
FROM category
WHERE budget = ?
ORDER BY name;

-- name: GetCategoryByName :one
SELECT id, budget, name
FROM category
WHERE budget = ? AND name = ?;

-- name: CreateCategoryGroup :one
INSERT INTO category_group (
  budget,
  name
) VALUES (
  ?,
  ?
)
RETURNING id;

-- name: UpdateCategoryGroup :execrows
UPDATE category_group
SET name = ?
WHERE id = ?;

-- name: GetOrCreateCategoryGroup :one
INSERT INTO category_group (
  budget,
  name
) VALUES (
  ?,
  ?
)
ON CONFLICT (budget, name) DO UPDATE SET name = excluded.name
RETURNING id;

-- name: DeleteCategoryGroup :exec
DELETE FROM category_group
WHERE id = ?;

-- name: ListCategoryGroups :many
SELECT id, budget, name
FROM category_group
WHERE budget = ?
ORDER BY name;

-- name: GetCategoryGroupByName :one
SELECT id, budget, name
FROM category_group
WHERE budget = ? AND name = ?;
