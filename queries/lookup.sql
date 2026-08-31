-- name: ListBudgets :many
SELECT id, name
FROM budget
ORDER BY id;

-- name: GetBudgetByName :one
SELECT id, name
FROM budget
WHERE name = ?;

-- name: GetAccountByName :one
SELECT id, budget, name
FROM account
WHERE budget = ? AND name = ?;

-- name: GetCategoryByName :one
SELECT id, budget, name
FROM category
WHERE budget = ? AND name = ?;

-- name: GetCategoryGroupByName :one
SELECT id, budget, name
FROM category_group
WHERE budget = ? AND name = ?;

-- name: GetPayeeByName :one
SELECT id, budget, name
FROM payee
WHERE budget = ? AND name = ?;

-- name: GetGoalByCategory :one
SELECT id, budget, type, start, "end", category, amount
FROM goal
WHERE budget = ? AND category = ?;
