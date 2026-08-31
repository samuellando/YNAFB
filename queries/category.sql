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