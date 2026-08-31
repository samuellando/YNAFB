-- name: ListAccountsByBudget :many
SELECT id, budget, name
FROM account
WHERE budget = ?
ORDER BY name;

-- name: ListCategoriesByBudget :many
SELECT id, budget, name
FROM category
WHERE budget = ?
ORDER BY name;

-- name: ListCategoryGroupsByBudget :many
SELECT id, budget, name
FROM category_group
WHERE budget = ?
ORDER BY name;

-- name: ListPayeesByBudget :many
SELECT id, budget, name
FROM payee
WHERE budget = ?
ORDER BY name;

-- name: ListGoalsByBudget :many
SELECT
  g.id,
  g.type,
  g.start,
  g."end",
  g.category AS category_id,
  c.name AS category_name,
  g.amount
FROM goal AS g
JOIN category AS c ON c.id = g.category
WHERE g.budget = ?
ORDER BY c.name;

-- name: ListAllocationsByBudget :many
SELECT
  a.id,
  a.month,
  a.category AS category_id,
  c.name AS category_name,
  a.amount
FROM allocation AS a
JOIN category AS c ON c.id = a.category
WHERE a.budget = ?
ORDER BY a.month, c.name;

-- name: ListTransactionsByBudget :many
SELECT
  t.id,
  t.date,
  a.name AS account_name,
  p.name AS payee_name,
  t.total_outflow,
  t.total_inflow,
  t.note
FROM "transaction" AS t
JOIN account AS a ON a.id = t.account
JOIN payee AS p ON p.id = t.payee
WHERE a.budget = ?
ORDER BY t.date DESC, t.id DESC;
