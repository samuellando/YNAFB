-- +goose up
-- The name a budget shows as to other members of an expense share.
-- ADD COLUMN with NOT NULL needs a compliant placeholder default for existing
-- rows; the UPDATE below immediately backfills every row from its budget name
-- (budget.name already satisfies LENGTH >= 3), so the placeholder is transient.
ALTER TABLE budget_expense_share
  ADD COLUMN display_name TEXT NOT NULL
  CHECK (LENGTH(display_name) >= 3) DEFAULT 'xxx';
UPDATE budget_expense_share
SET
  display_name = (
    SELECT
      name
    FROM
      budget
    WHERE
      budget.id = budget_expense_share.budget_id
  );

-- +goose down
ALTER TABLE budget_expense_share DROP COLUMN display_name;
