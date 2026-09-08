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
SELECT
  login.*
FROM
  login
WHERE
  login.username = ?;

-- name: ListLogins :many
SELECT
  login.*
FROM
  login
ORDER BY
  login.id;

-- name: UpdateLoginPassword :one
UPDATE login
SET
  password = ?
WHERE
  login.id = @id
RETURNING *;

-- name: DeleteLogin :exec
DELETE FROM login
WHERE
  login.id = @id;