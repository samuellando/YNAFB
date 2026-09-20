-- +goose NO TRANSACTION
-- +goose up
-- Allow stored splits of zero so members who are not paying can have an
-- explicit $0 split (previously the CHECK forced one side strictly positive,
-- and omitting the row would synthesize a default share instead).
PRAGMA foreign_keys = OFF;
PRAGMA legacy_alter_table = ON;

ALTER TABLE expense_share_trx_split
RENAME TO expense_share_trx_split_old;

CREATE TABLE expense_share_trx_split (
  id INTEGER PRIMARY KEY,
  expense_share_trx_id INTEGER NOT NULL REFERENCES expense_share_trx (id) ON DELETE CASCADE,
  expense_share_id INTEGER NOT NULL REFERENCES expense_share (id) ON DELETE CASCADE,
  budget_id INTEGER references budget (id) ON DELETE SET NULL,
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
  UNIQUE(budget_id, expense_share_trx_id),
  UNIQUE (budget_id, id),
  FOREIGN KEY (budget_id, expense_share_id) REFERENCES budget_expense_share (budget_id, expense_share_id) ON DELETE SET NULL
);

INSERT INTO
  expense_share_trx_split (
    id,
    expense_share_trx_id,
    expense_share_id,
    budget_id,
    split_inflow,
    split_outflow
  )
SELECT
  id,
  expense_share_trx_id,
  expense_share_id,
  budget_id,
  split_inflow,
  split_outflow
FROM
  expense_share_trx_split_old;

DROP TABLE expense_share_trx_split_old;

PRAGMA legacy_alter_table = OFF;
PRAGMA foreign_keys = ON;

-- +goose down
PRAGMA foreign_keys = OFF;
PRAGMA legacy_alter_table = ON;

ALTER TABLE expense_share_trx_split
RENAME TO expense_share_trx_split_old;

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

-- DATA LOSS for zero splits, which the old CHECK rejects.
INSERT INTO
  expense_share_trx_split (
    id,
    expense_share_trx_id,
    expense_share_id,
    budget_id,
    split_inflow,
    split_outflow
  )
SELECT
  id,
  expense_share_trx_id,
  expense_share_id,
  budget_id,
  split_inflow,
  split_outflow
FROM
  expense_share_trx_split_old
WHERE
  split_inflow > 0
  OR split_outflow > 0;

DROP TABLE expense_share_trx_split_old;

PRAGMA legacy_alter_table = OFF;
PRAGMA foreign_keys = ON;
