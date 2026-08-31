-- name: DeleteBudget :exec
DELETE FROM budget
WHERE id = ?;

-- name: DeleteAccount :exec
DELETE FROM account
WHERE id = ?;

-- name: DeleteCategory :exec
DELETE FROM category
WHERE id = ?;

-- name: DeleteCategoryGroup :exec
DELETE FROM category_group
WHERE id = ?;

-- name: DeletePayee :exec
DELETE FROM payee
WHERE id = ?;

-- name: DeleteGoal :exec
DELETE FROM goal
WHERE id = ?;

-- name: DeleteAllocationByCategoryAndMonth :exec
DELETE FROM allocation
WHERE budget = ? AND category = ? AND month = ?;

-- name: DeleteTransaction :exec
DELETE FROM "transaction"
WHERE id = ?;

-- name: DeleteTransactionCategory :exec
DELETE FROM transaction_category
WHERE id = ?;

-- name: DeletePayeeDefaultCategory :exec
DELETE FROM payee_default_category
WHERE id = ?;

-- name: DeletePayeeDefaultCategoriesByPayee :exec
DELETE FROM payee_default_category
WHERE payee = ?;

-- name: DeleteTransactionCategoriesByTransaction :exec
DELETE FROM transaction_category
WHERE "transaction" = ?;