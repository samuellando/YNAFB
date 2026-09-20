-- +goose NO TRANSACTION
-- +goose Up
-- Deleting a category group must orphan its categories (category_group_id =
-- NULL) instead of deleting them. SQLite cannot alter a foreign key action in
-- place, so the category table is rebuilt with ON DELETE SET NULL.
-- NOTE: two overlapping keys on purpose. For composite keys SQLite's SET NULL
-- nulls every column of the child key, which would violate
-- category.budget_id's NOT NULL constraint on every group delete — hence the
-- single-column key carrying the SET NULL action. The retained composite key
-- (default NO ACTION) keeps scoping group assignment to the category's budget
-- (verified: group delete nulls, cross-budget assignment still fails).
-- PRAGMA foreign_keys is a no-op inside a transaction, hence NO TRANSACTION
-- above: enforcement is disabled for the rebuild (dropping the old table with
-- enforcement on would fire ON DELETE CASCADE on dependent rows) and
-- re-enabled afterwards. legacy_alter_table is enabled so the renames below
-- do not touch the budget_month_categories view that reads category;
-- everything is name-identical when done. Migrations run sequentially before
-- the pool is shared, so the pragmas hold for every statement here.
PRAGMA foreign_keys = OFF;
PRAGMA legacy_alter_table = ON;

ALTER TABLE category RENAME TO category_old;

CREATE TABLE category (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
  name TEXT NOT NULL CHECK (LENGTH(name) >= 3),
  category_group_id INTEGER REFERENCES category_group (id) ON DELETE SET NULL,
  CHECK (lower(name) IS NOT 'income'),
  UNIQUE (budget_id, id),
  UNIQUE (budget_id, name),
  FOREIGN KEY (budget_id, category_group_id) REFERENCES category_group (budget_id, id)
);

INSERT INTO category (id, budget_id, name, category_group_id)
  SELECT id, budget_id, name, category_group_id FROM category_old;

DROP TABLE category_old;

PRAGMA legacy_alter_table = OFF;
PRAGMA foreign_keys = ON;

-- +goose NO TRANSACTION
-- +goose Down
-- Restore ON DELETE CASCADE (same rebuild in reverse).
PRAGMA foreign_keys = OFF;
PRAGMA legacy_alter_table = ON;

ALTER TABLE category RENAME TO category_old;

CREATE TABLE category (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
  name TEXT NOT NULL CHECK (LENGTH(name) >= 3),
  category_group_id INTEGER,
  CHECK (lower(name) IS NOT 'income'),
  UNIQUE (budget_id, id),
  UNIQUE (budget_id, name),
  FOREIGN KEY (budget_id, category_group_id) REFERENCES category_group (budget_id, id) ON DELETE CASCADE
);

INSERT INTO category (id, budget_id, name, category_group_id)
  SELECT id, budget_id, name, category_group_id FROM category_old;

DROP TABLE category_old;

PRAGMA legacy_alter_table = OFF;
PRAGMA foreign_keys = ON;
