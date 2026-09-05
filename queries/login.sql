-- name: CreateLogin :one
INSERT INTO login (
  username,
  password
) VALUES (
  ?,
  ?
)
RETURNING *;

-- name: GetLoginByUsername :one
SELECT *
FROM login
WHERE username = ?;

-- name: ListLogins :many
SELECT *
FROM login
ORDER BY id;

-- name: UpdateLoginPassword :execrows
UPDATE login
SET password = ?
WHERE id = ?;

-- name: DeleteLogin :exec
DELETE FROM login
WHERE id = ?;