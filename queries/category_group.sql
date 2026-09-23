-- name: CreateCategoryGroup :one
INSERT INTO
  category_group (budget_id, name)
SELECT
  b.id, ?
FROM
  budget AS b
WHERE
  b.id = @budget_id AND b.login_id = @login_id
RETURNING
  id;

-- name: UpdateCategoryGroup :one
UPDATE category_group
SET
  name = ?
WHERE
  category_group.id = @id
  AND category_group.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id AND b.login_id = @login_id
  )
RETURNING *;

-- name: DeleteCategoryGroup :exec
DELETE FROM category_group
WHERE
  category_group.id = @id
  AND category_group.budget_id IN (
    SELECT
      b.id
    FROM
      budget AS b
    WHERE
      b.id = @budget_id AND b.login_id = @login_id
  );

-- name: ListCategoryGroups :many
SELECT
  cg.id,
  cg.budget_id,
  cg.name
FROM
  category_group AS cg
  JOIN budget AS b ON cg.budget_id = b.id
WHERE
  b.login_id = @login_id AND cg.budget_id = @budget_id
ORDER BY
  cg.name;

-- name: GetCategoryGroup :one
SELECT
  cg.id,
  cg.budget_id,
  cg.name
FROM
  category_group AS cg
  JOIN budget AS b ON cg.budget_id = b.id
WHERE
  b.login_id = @login_id
  AND cg.budget_id = @budget_id
  AND cg.id = @id;
