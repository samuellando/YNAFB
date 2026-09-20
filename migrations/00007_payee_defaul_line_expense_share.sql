-- +goose NO TRANSACTION
-- +goose up
--
PRAGMA foreign_keys = OFF;
PRAGMA legacy_alter_table = ON;

ALTER TABLE payee_default_line
RENAME TO payee_default_line_old;

CREATE TABLE payee_default_line (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER NOT NULL,
  payee_id INTEGER NOT NULL,
  dest_account_id INTEGER,
  category_id INTEGER,
  expense_share_id INTEGER,
  income BOOL NOT NULL DEFAULT false,
  percent INTEGER NOT NULL,
  -- XOR, can only be one of category (spend), transfer, or income
  CHECK (
    (
      CASE
        WHEN dest_account_id IS NOT NULL THEN 1
        ELSE 0
      END + CASE
        WHEN category_id IS NOT NULL THEN 1
        ELSE 0
      END + CASE
        WHEN income THEN 1
        ELSE 0
      END + CASE
        WHEN expense_share_id IS NOT NULL THEN 1
        ELSE 0
      END
    ) = 1
  ),
  CHECK (
    percent > 0
    AND percent <= 100
  ),
  FOREIGN KEY (budget_id, payee_id) REFERENCES payee (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, dest_account_id) REFERENCES account (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, expense_share_id) REFERENCES budget_expense_share (budget_id, expense_share_id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, category_id) REFERENCES category (budget_id, id) ON DELETE CASCADE
);

INSERT INTO
  payee_default_line (
    id,
    budget_id,
    payee_id,
    dest_account_id,
    category_id,
    expense_share_id,
    income,
  percent
  )
SELECT
  id,
  budget_id,
  payee_id,
  dest_account_id,
  category_id,
  NULL,
  income,
  percent
FROM
  payee_default_line_old;

DROP TABLE payee_default_line_old;

PRAGMA legacy_alter_table = OFF;
PRAGMA foreign_keys = ON;

-- +goose down

PRAGMA foreign_keys = OFF;
PRAGMA legacy_alter_table = ON;

ALTER TABLE payee_default_line
RENAME TO payee_default_line_old;

CREATE TABLE payee_default_line (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER NOT NULL,
  payee_id INTEGER NOT NULL,
  dest_account_id INTEGER,
  category_id INTEGER,
  income BOOL NOT NULL DEFAULT false,
  percent INTEGER NOT NULL,
  -- XOR, can only be one of category (spend), transfer, or income
  CHECK (
    (
      CASE
        WHEN dest_account_id IS NOT NULL THEN 1
        ELSE 0
      END + CASE
        WHEN category_id IS NOT NULL THEN 1
        ELSE 0
      END + CASE
        WHEN income THEN 1
        ELSE 0
      END
    ) = 1
  ),
  CHECK (
    percent > 0
    AND percent <= 100
  ),
  FOREIGN KEY (budget_id, payee_id) REFERENCES payee (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, dest_account_id) REFERENCES account (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, category_id) REFERENCES category (budget_id, id) ON DELETE CASCADE
);

INSERT INTO
  payee_default_line (
    id,
    budget_id,
    payee_id,
    dest_account_id,
    category_id,
    income,
    percent
  )
SELECT
  id,
  budget_id,
  payee_id,
  dest_account_id,
  category_id,
  income,
  percent
FROM
  payee_default_line_old;

DROP TABLE payee_default_line_old;

PRAGMA legacy_alter_table = OFF;
PRAGMA foreign_keys = ON;
