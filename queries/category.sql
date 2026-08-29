-- name: CreateCategory :one
INSERT INTO category (
  budget,
  name
) VALUES (
  ?,
  ?
)
RETURNING *;
