-- +goose NO TRANSACTION
-- +goose up
CREATE TABLE expense_share (
  id INTEGER PRIMARY KEY
);

-- Share codes for the expense share so other users can add it.
CREATE TABLE expense_share_codes (
  id INTEGER PRIMARY KEY,
  expense_share_id integer NOT NULL REFERENCES expense_share (id) ON DELETE CASCADE,
  code TEXT NOT NULL CHECK (LENGTH(code) >= 10),
  created UNIX_EPOCH_INTEGER NOT NULL
);

-- One or more users in an expense share
CREATE TABLE budget_expense_share (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL CHECK (LENGTH(name) >= 3),
  expense_share_id integer NOT NULL REFERENCES expense_share (id) ON DELETE CASCADE,
  budget_id integer NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
  UNIQUE (budget_id, expense_share_id),
  UNIQUE (budget_id, name)
);

-- A user's trx was categorized with at least one line pointing to this expense share
-- The total is a trx total, and requested is the amount they marked for this expense share
CREATE TABLE expense_share_trx (
  id INTEGER PRIMARY KEY,
  expense_share_id INTEGER NOT NULL REFERENCES expense_share (id) ON DELETE CASCADE,
  -- Reference back to the source
  budget_id INTEGER REFERENCES budget (id) ON DELETE SET NULL,
  trx_id INTEGER REFERENCES trx (id) ON DELETE SET NULL,
  -- Copy of the source information
  payee_name TEXT NOT NULL,
  date UNIX_EPOCH_INTEGER NOT NULL,
  total_outflow INTEGER NOT NULL DEFAULT 0,
  total_inflow INTEGER NOT NULL DEFAULT 0,
  requested_outflow INTEGER NOT NULL DEFAULT 0,
  requested_inflow INTEGER NOT NULL DEFAULT 0,
  note TEXT NOT NULL DEFAULT '',
  CHECK (
    (
      total_inflow > 0
      AND total_outflow = 0
    )
    OR (
      total_outflow > 0
      AND total_inflow = 0
    )
  ),
  CHECK (
    (
      requested_inflow > 0
      AND requested_outflow = 0
    )
    OR (
      requested_outflow > 0
      AND requested_inflow = 0
    )
  ),
  CHECK (
    (
      total_outflow > 0
      AND requested_outflow > 0
    )
    OR (
      total_inflow > 0
      AND requested_inflow > 0
    )
  ),
  CHECK (
      requested_outflow <= total_outflow
      AND requested_inflow <= total_inflow
  ),
  FOREIGN KEY (budget_id, trx_id) REFERENCES trx (budget_id, id) ON DELETE SET NULL,
  FOREIGN KEY (budget_id, expense_share_id) REFERENCES budget_expense_share (budget_id, expense_share_id) ON DELETE SET NULL,
  UNIQUE (budget_id, expense_share_id, trx_id)
);

-- The portion of the requested amount other users other than the requestor are paying
-- should total the requested amount on the shared transaction
CREATE TABLE expense_share_trx_split (
  id INTEGER PRIMARY KEY,
  expense_share_trx_id INTEGER NOT NULL REFERENCES expense_share_trx (id) ON DELETE CASCADE,
  expense_share_id INTEGER NOT NULL REFERENCES expense_share (id) ON DELETE CASCADE,
  budget_id INTEGER references budget (id) ON DELETE SET NULL,
  split_inflow INTEGER NOT NULL DEFAULT 0,
  split_outflow INTEGER NOT NULL DEFAULT 0 CHECK (
    (
      split_inflow > 0
      AND split_outflow = 0
    )
    OR (
      split_outflow > 0
      AND split_inflow = 0
    )
  ),
  UNIQUE(budget_id, expense_share_trx_id),
  UNIQUE (budget_id, id),
  FOREIGN KEY (budget_id, expense_share_id) REFERENCES budget_expense_share (budget_id, expense_share_id) ON DELETE SET NULL 
);

-- How the user categorizes their split of a shared transaction in their budget
CREATE TABLE expense_share_trx_split_line (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER NOT NULL references budget (id) ON DELETE CASCADE,
  expense_share_trx_split_id INTEGER NOT NULL REFERENCES expense_share_trx_split (id) ON DELETE CASCADE,
  dest_account_id INTEGER,
  category_id INTEGER,
  outflow INTEGER NOT NULL DEFAULT 0,
  inflow INTEGER NOT NULL DEFAULT 0,
  -- XOR, can only be one of category (spend), transfer, or income
  CHECK (
    (
      CASE
        WHEN dest_account_id IS NOT NULL THEN 1
        ELSE 0
      END + CASE
        WHEN category_id IS NOT NULL THEN 1
        ELSE 0
      END  
    ) = 1
  ),
  CHECK (
    (
      inflow > 0
      AND outflow = 0
    )
    OR (
      outflow > 0
      AND inflow = 0
    )
  ),
  FOREIGN KEY (budget_id, expense_share_trx_split_id) REFERENCES expense_share_trx_split (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, dest_account_id) REFERENCES account (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, category_id) REFERENCES category (budget_id, id) ON DELETE CASCADE
);

PRAGMA foreign_keys = OFF;
PRAGMA legacy_alter_table = ON;

ALTER TABLE trx_line
RENAME TO trx_line_old;

CREATE TABLE trx_line (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER NOT NULL,
  trx_id INTEGER NOT NULL,
  dest_account_id INTEGER,
  category_id INTEGER,
  expense_share_id INTEGER,
  income BOOL NOT NULL DEFAULT false,
  outflow INTEGER NOT NULL DEFAULT 0,
  inflow INTEGER NOT NULL DEFAULT 0,
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
    NOT INCOME
    OR (
      outflow = 0
      and inflow > 0
    )
  ),
  CHECK (
    (
      inflow > 0
      AND outflow = 0
    )
    OR (
      outflow > 0
      AND inflow = 0
    )
  ),
  FOREIGN KEY (budget_id, trx_id) REFERENCES trx (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, dest_account_id) REFERENCES account (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, expense_share_id) REFERENCES budget_expense_share (budget_id, expense_share_id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, category_id) REFERENCES category (budget_id, id) ON DELETE CASCADE
);

INSERT INTO
  trx_line (
    id,
    budget_id,
    trx_id,
    dest_account_id,
    category_id,
    expense_share_id,
    income,
    outflow,
    inflow
  )
SELECT
  id,
  budget_id,
  trx_id,
  dest_account_id,
  category_id,
  NULL,
  income,
  outflow,
  inflow
FROM
  trx_line_old;

DROP TABLE trx_line_old;

PRAGMA legacy_alter_table = OFF;
PRAGMA foreign_keys = ON;

-- +goose down

PRAGMA foreign_keys = OFF;
PRAGMA legacy_alter_table = ON;

DROP TABLE expense_share_trx_split_line;
DROP TABLE expense_share_trx_split;
DROP TABLE expense_share_trx;
DROP TABLE expense_share_codes;
DROP TABLE budget_expense_share;
DROP TABLE expense_share;

ALTER TABLE trx_line
RENAME TO trx_line_old;

CREATE TABLE trx_line (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER NOT NULL,
  trx_id INTEGER NOT NULL,
  dest_account_id INTEGER,
  category_id INTEGER,
  income BOOL NOT NULL DEFAULT false,
  outflow INTEGER NOT NULL DEFAULT 0,
  inflow INTEGER NOT NULL DEFAULT 0,
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
    NOT INCOME
    OR (
      outflow = 0
      and inflow > 0
    )
  ),
  CHECK (
    (
      inflow > 0
      AND outflow = 0
    )
    OR (
      outflow > 0
      AND inflow = 0
    )
  ),
  FOREIGN KEY (budget_id, trx_id) REFERENCES trx (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, dest_account_id) REFERENCES account (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, category_id) REFERENCES category (budget_id, id) ON DELETE CASCADE
);

-- DATA LOSS for the expense_share_id
INSERT INTO
  trx_line (
    id,
    budget_id,
    trx_id,
    dest_account_id,
    category_id,
    income,
    outflow,
    inflow
  )
SELECT
  id,
  budget_id,
  trx_id,
  dest_account_id,
  category_id,
  income,
  outflow,
  inflow
FROM
  trx_line_old;

DROP TABLE trx_line_old;

PRAGMA legacy_alter_table = OFF;
PRAGMA foreign_keys = ON;
