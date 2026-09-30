-- +goose NO TRANSACTION
-- +goose up

-- Expanse shares, a golbal entry that can be shared accross budgets and accross users
CREATE TABLE expense_share (
  id INTEGER PRIMARY KEY,
  default_name TEXT NOT NULL CHECK (LENGTH(default_name) >= 3)
);

-- Share codes for the expense share so other users can add it.
CREATE TABLE expense_share_code (
  id INTEGER PRIMARY KEY,
  expense_share_id integer NOT NULL REFERENCES expense_share (id) ON DELETE CASCADE,
  code TEXT NOT NULL CHECK (LENGTH(code) >= 10),
  expires UNIX_EPOCH_INTEGER NOT NULL
);

-- One or more users in an expense share
CREATE TABLE budget_expense_share (
  id INTEGER PRIMARY KEY,
  -- How this expnse share is displayed for the user
  name TEXT NOT NULL CHECK (LENGTH(name) >= 3),
  -- How this user is displayed to other users
  display_name TEXT NOT NULL CHECK (LENGTH(display_name) >= 3),
  -- Deleteing the parent expense share would delete if for all users
  expense_share_id integer NOT NULL REFERENCES expense_share (id) ON DELETE CASCADE,
  -- The budget within the expense share
  budget_id integer NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
  -- A budget can only join an expense share once
  UNIQUE (budget_id, expense_share_id),
  -- A budget must have a unique name for all it's expnse share to avoid confusion
  UNIQUE (budget_id, name),
  -- All display names in expense shares must be unique to avoid confusion
  UNIQUE (expense_share_id, display_name)
);

-- A transaction a user has published to an expense share.
-- Handles the user leaving the expense shares by taking snapshots.
-- Takes a snapshot of the in
CREATE TABLE expense_share_trx (
  id INTEGER PRIMARY KEY,
  -- A reference to the source expense share
  expense_share_id INTEGER NOT NULL REFERENCES expense_share (id) ON DELETE CASCADE,
  -- Reference back to the source, nullable if the source is deleted or leaves
  budget_expense_share_id INTEGER REFERENCES budget_expense_share (id) ON DELETE SET NULL,
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
  -- The transaction must either be an inflow or outflow
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
  -- The amount the user is requesting other's to pay
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
  -- the signs must match
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
  -- Can only request less or equal to the amount
  CHECK (
      requested_outflow <= total_outflow
      AND requested_inflow <= total_inflow
  ),
  -- The budget must actual be part of the expense share
  FOREIGN KEY (budget_id, budget_expense_share_id) REFERENCES budget_expense_share (budget_id, id) ON DELETE SET NULL,
  FOREIGN KEY (expense_share_id, budget_expense_share_id) REFERENCES budget_expense_share (expense_share_id, id),
  -- The transaction must belong to the budget
  FOREIGN KEY (budget_id, trx_id) REFERENCES trx (budget_id, id) ON DELETE SET NULL
);

-- The portion of the requested amount other users other than the requestor are paying
-- should total the requested amount on the shared transaction, checked in the domain layer
CREATE TABLE expense_share_trx_split (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER REFERENCES budget (id) ON DELETE SET NULL,
  expense_share_id INTEGER NOT NULL REFERENCES expense_share (id) ON DELETE CASCADE,
  expense_share_trx_id INTEGER NOT NULL REFERENCES expense_share_trx (id) ON DELETE CASCADE,
  split_inflow INTEGER NOT NULL DEFAULT 0,
  split_outflow INTEGER NOT NULL DEFAULT 0 CHECK (
    (
      split_inflow >= 0
      AND split_outflow = 0
    )
    OR (
      split_outflow >= 0
      AND split_inflow = 0
    )
  ),
  -- The budget must be part of the expense share
  FOREIGN KEY (budget_id, expense_share_id) REFERENCES budget_expense_share (budget_id, expense_share_id) ON DELETE SET NULL,
  -- The transaction must match with the expense share
  FOREIGN KEY (expense_share_trx_id, expense_share_id) REFERENCES expense_share_trx (id, expense_share_id) ON DELETE CASCADE
);

-- How the user categorizes their split of a shared transaction in their budget
-- Not share with other users
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
  -- All the values must belong to the budget
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

DROP TABLE expense_share_trx_split_line;
DROP TABLE expense_share_trx_split;
DROP TABLE expense_share_trx;
DROP TABLE expense_share_code;
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
