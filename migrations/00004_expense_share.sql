-- +goose NO TRANSACTION
-- +goose up

-- Expense shares, a global entry that can be shared across budgets and across users
-- Ownerless: no login_id/budget owner.
CREATE TABLE expense_share (
  id INTEGER PRIMARY KEY,
  default_name TEXT NOT NULL CHECK (LENGTH(default_name) >= 3)
);

-- Share codes for the expense share so other users can add it.
CREATE TABLE expense_share_code (
  id INTEGER PRIMARY KEY,
  expense_share_id integer NOT NULL REFERENCES expense_share (id) ON DELETE CASCADE,
  code TEXT NOT NULL UNIQUE CHECK (LENGTH(TRIM(code)) >= 10),
  expires UNIX_EPOCH_INTEGER NOT NULL
);

-- One or more users in an expense share
CREATE TABLE budget_expense_share (
  id INTEGER PRIMARY KEY,
  -- How this expense share is displayed for the user
  name TEXT NOT NULL CHECK (LENGTH(name) >= 3),
  -- How this user is displayed to other users
  display_name TEXT NOT NULL CHECK (LENGTH(display_name) >= 3),
  -- Deleting the parent expense share deletes it for all users
  expense_share_id integer NOT NULL REFERENCES expense_share (id) ON DELETE CASCADE,
  -- The budget within the expense share
  budget_id integer NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
  -- A budget can only join an expense share once
  UNIQUE (budget_id, expense_share_id),
  -- A budget must have a unique name for all its expense shares to avoid confusion
  UNIQUE (budget_id, name),
  -- All display names in expense shares must be unique to avoid confusion
  UNIQUE (expense_share_id, display_name),
  -- Composite keys for FK scoping (cf. account/category UNIQUE(budget_id, id))
  UNIQUE (budget_id, id),
  UNIQUE (expense_share_id, id)
);

-- How the counterparty categorizes their share of a shared transaction.
-- Cross-budget by design
CREATE TABLE expense_share_trx_split_line (
  id INTEGER PRIMARY KEY,
  -- Owner: the tagged counterparty categorizing in their own budget
  budget_id INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
  expense_share_id INTEGER NOT NULL,
  -- Source: the original transaction being categorized (other budget)
  source_budget_id INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
  source_trx_id INTEGER NOT NULL,
  -- Optional link to the exact split line being categorized
  source_trx_line_id INTEGER REFERENCES trx_line (id) ON DELETE CASCADE,
  category_id INTEGER,
  outflow INTEGER NOT NULL DEFAULT 0,
  inflow INTEGER NOT NULL DEFAULT 0,
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
  -- Cross-budget by design: owner and source must differ
  CHECK (budget_id != source_budget_id),
  -- Owner and source must both be members of the same expense share
  FOREIGN KEY (budget_id, expense_share_id) REFERENCES budget_expense_share (budget_id, expense_share_id) ON DELETE CASCADE,
  FOREIGN KEY (source_budget_id, expense_share_id) REFERENCES budget_expense_share (budget_id, expense_share_id) ON DELETE CASCADE,
  -- Live reference: source delete cascades categorizations (no snapshot)
  FOREIGN KEY (source_budget_id, source_trx_id) REFERENCES trx (budget_id, id) ON DELETE CASCADE,
  -- Categorization targets belong to the owner's budget
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
  -- expense share qualifier: which share this split/settlement belongs to
  expense_share_id INTEGER,
  -- split with another member of an expense share
  split_budget_id INTEGER,
  -- settlement with another member of an expense share
  dest_budget_id INTEGER,
  income BOOL NOT NULL DEFAULT false CHECK (income IN (0, 1)),
  outflow INTEGER NOT NULL DEFAULT 0,
  inflow INTEGER NOT NULL DEFAULT 0,
  -- XOR, can only be one of category (spend), transfer, income, split line, or settlement line
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
        WHEN split_budget_id IS NOT NULL THEN 1
        ELSE 0
      END + CASE
        WHEN dest_budget_id IS NOT NULL THEN 1
        ELSE 0
      END
    ) = 1
  ),
  -- Split/settlement lines must name their share; other lines must not
  CHECK (
    (split_budget_id IS NULL AND dest_budget_id IS NULL)
    OR expense_share_id IS NOT NULL
  ),
  CHECK (
    expense_share_id IS NULL
    OR (split_budget_id IS NOT NULL OR dest_budget_id IS NOT NULL)
  ),
  -- No self-splits or self-settlements (NULL passes, non-NULL must differ)
  CHECK (split_budget_id != budget_id),
  CHECK (dest_budget_id != budget_id),
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
  FOREIGN KEY (budget_id, category_id) REFERENCES category (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, expense_share_id) REFERENCES budget_expense_share (budget_id, expense_share_id) ON DELETE CASCADE,
  FOREIGN KEY (split_budget_id, expense_share_id) REFERENCES budget_expense_share (budget_id, expense_share_id) ON DELETE CASCADE,
  FOREIGN KEY (dest_budget_id, expense_share_id) REFERENCES budget_expense_share (budget_id, expense_share_id) ON DELETE CASCADE
);

INSERT INTO
  trx_line (
    id,
    budget_id,
    trx_id,
    dest_account_id,
    category_id,
    expense_share_id,
    split_budget_id,
    dest_budget_id,
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
  NULL,
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
  -- expense share qualifier: which share this split/settlement belongs to
  expense_share_id INTEGER,
  -- split with another member of an expense share
  split_budget_id INTEGER,
  -- settlement with another member of an expense share
  dest_budget_id INTEGER,
  income BOOL NOT NULL DEFAULT false CHECK (income IN (0, 1)),
  percent INTEGER NOT NULL,
  -- XOR, can only be one of category (spend), transfer, income, split line, or settlement line
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
        WHEN split_budget_id IS NOT NULL THEN 1
        ELSE 0
      END + CASE
        WHEN dest_budget_id IS NOT NULL THEN 1
        ELSE 0
      END
    ) = 1
  ),
  -- Split/settlement lines must name their share; other lines must not
  CHECK (
    (split_budget_id IS NULL AND dest_budget_id IS NULL)
    OR expense_share_id IS NOT NULL
  ),
  CHECK (
    expense_share_id IS NULL
    OR (split_budget_id IS NOT NULL OR dest_budget_id IS NOT NULL)
  ),
  -- No self-splits or self-settlements (NULL passes, non-NULL must differ)
  CHECK (split_budget_id != budget_id),
  CHECK (dest_budget_id != budget_id),
  CHECK (
    percent > 0
    AND percent <= 100
  ),
  FOREIGN KEY (budget_id, payee_id) REFERENCES payee (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, dest_account_id) REFERENCES account (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, category_id) REFERENCES category (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, expense_share_id) REFERENCES budget_expense_share (budget_id, expense_share_id) ON DELETE CASCADE,
  FOREIGN KEY (split_budget_id, expense_share_id) REFERENCES budget_expense_share (budget_id, expense_share_id) ON DELETE CASCADE,
  FOREIGN KEY (dest_budget_id, expense_share_id) REFERENCES budget_expense_share (budget_id, expense_share_id) ON DELETE CASCADE
);

INSERT INTO
  payee_default_line (
    id,
    budget_id,
    payee_id,
    dest_account_id,
    category_id,
    expense_share_id,
    split_budget_id,
    dest_budget_id,
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
  NULL,
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
  income BOOL NOT NULL DEFAULT false CHECK (income IN (0, 1)),
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
  income BOOL NOT NULL DEFAULT false CHECK (income IN (0, 1)),
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
