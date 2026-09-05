-- +goose up
CREATE TABLE login (
  id INTEGER PRIMARY KEY,
  username TEXT NOT NULL UNIQUE CHECK (LENGTH(username) >= 3),
  password TEXT NOT NULL
);

CREATE TABLE budget (
  id INTEGER PRIMARY KEY,
  login_id INTEGER NOT NULL REFERENCES login (id) ON DELETE CASCADE,
  name TEXT NOT NULL CHECK (LENGTH(name) >= 3),
  UNIQUE (login_id, name)
);

CREATE TABLE account (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
  name TEXT NOT NULL CHECK (LENGTH(name) >= 3),
  UNIQUE (budget_id, id),
  UNIQUE (budget_id, name)
);

CREATE TABLE payee (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
  name TEXT NOT NULL CHECK (LENGTH(name) >= 3),
  UNIQUE (budget_id, id),
  UNIQUE (budget_id, name)
);

CREATE TABLE trx (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER NOT NULL,
  account_id INTEGER NOT NULL,
  payee_id INTEGER NOT NULL,
  date UNIX_EPOCH_INTEGER NOT NULL,
  total_outflow INTEGER NOT NULL DEFAULT 0,
  total_inflow INTEGER NOT NULL DEFAULT 0,
  note TEXT NOT NULL DEFAULT '',
  UNIQUE (budget_id, id),
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
  FOREIGN KEY (budget_id, account_id) REFERENCES account (budget_id, id) ON DELETE CASCADE,
  FOREIGN KEY (budget_id, payee_id) REFERENCES payee (budget_id, id) ON DELETE CASCADE
);

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

CREATE TABLE category_group (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
  name TEXT NOT NULL CHECK (LENGTH(name) >= 3),
  UNIQUE (budget_id, id),
  UNIQUE (budget_id, name)
);

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

CREATE TABLE allocation (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER NOT NULL,
  category_id INTEGER NOT NULL,
  month UNIX_EPOCH_INTEGER NOT NULL,
  amount INTEGER NOT NULL,
  UNIQUE (budget_id, category_id, month),
  FOREIGN KEY (budget_id, category_id) REFERENCES category (budget_id, id) ON DELETE CASCADE
);

CREATE TABLE reconciliation (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER NOT NULL,
  account_id INTEGER NOT NULL,
  trx_id INTEGER REFERENCES trx (id) ON DELETE CASCADE,
  trx_line_id INTEGER REFERENCES trx_line (id) ON DELETE CASCADE,
  CHECK (
    (
      CASE
        WHEN trx_id IS NOT NULL THEN 1
        ELSE 0
      END + CASE
        WHEN trx_line_id IS NOT NULL THEN 1
        ELSE 0
      END
    ) = 1
  ),
  FOREIGN KEY (budget_id, account_id) REFERENCES account (budget_id, id) ON DELETE CASCADE
);

CREATE TABLE goal (
  id INTEGER PRIMARY KEY,
  budget_id INTEGER NOT NULL,
  category_id INTEGER NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('monthly', 'save', 'refill')),
  start_date UNIX_EPOCH_INTEGER NOT NULL,
  end_date UNIX_EPOCH_INTEGER,
  amount INTEGER NOT NULL,
  UNIQUE (budget_id, category_id),
  CHECK (
    type IN ('monthly', 'refill')
    OR end_date IS NOT NULL
  ),
  CHECK (
    end_date IS NULL
    OR end_date >= start_date
  ),
  CHECK (amount > 0),
  FOREIGN KEY (budget_id, category_id) REFERENCES category (budget_id, id) ON DELETE CASCADE
);

-- +goose down
DROP TABLE goal;

DROP TABLE reconciliation;

DROP TABLE allocation;

DROP TABLE payee_default_line;

DROP TABLE category_group;

DROP TABLE category;

DROP TABLE trx_line;

DROP TABLE trx;

DROP TABLE payee;

DROP TABLE account;

DROP TABLE budget;

DROP TABLE login;
