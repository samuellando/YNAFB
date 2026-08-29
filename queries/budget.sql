-- name: CreateBudget :one
INSERT INTO budget (
  name
) VALUES (
  ?
)
RETURNING *;
