-- name: CreatePayeeDefaultCategory :one
INSERT INTO payee_default_category (
  payee,
  other_account,
  category,
  income,
  percent
) VALUES (
  ?,
  ?,
  ?,
  ?,
  ?
)
RETURNING *;

-- name: UpdatePayeeDefaultCategory :execrows
UPDATE payee_default_category
SET payee = ?, other_account = ?, category = ?, income = ?, percent = ?
WHERE id = ?;

-- name: DeletePayeeDefaultCategory :exec
DELETE FROM payee_default_category
WHERE id = ?;

-- name: DeletePayeeDefaultCategoriesByPayee :exec
DELETE FROM payee_default_category
WHERE payee = ?;

-- name: ListPayeeDefaultCategoriesByPayee :many
SELECT
  pdc.id,
  pdc.payee,
  pdc.other_account,
  ao.name AS other_account_name,
  pdc.category,
  c.name AS category_name,
  pdc.income,
  pdc.percent
FROM payee_default_category AS pdc
LEFT JOIN account AS ao ON ao.id = pdc.other_account
LEFT JOIN category AS c ON c.id = pdc.category
WHERE pdc.payee = ?
ORDER BY pdc.id;
