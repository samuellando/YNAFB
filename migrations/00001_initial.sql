-- +goose Up
CREATE TABLE IF NOT EXISTS budget (
    id INTEGER PRIMARY KEY,
    name string NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS allocation (
    id INTEGER PRIMARY KEY,
    budget INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    category INTEGER NOT NULL REFERENCES category (id) ON DELETE CASCADE,
    amount INTEGER NOT NULL,
    UNIQUE (budget, category)
);

CREATE TABLE IF NOT EXISTS account (
    id INTEGER PRIMARY KEY,
    budget INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    name STRING NOT NULL,
    UNIQUE (budget, name)
);

CREATE TABLE IF NOT EXISTS "transaction" (
    id INTEGER PRIMARY KEY,
    date DATETIME NOT NULL,
    account INTEGER NOT NULL REFERENCES account (id) ON DELETE CASCADE,
    payee INTEGER NOT NULL REFERENCES payee (id) ON DELETE CASCADE,
    total_outflow INTEGER NOT NULL DEFAULT 0,
    total_inflow INTEGER NOT NULL DEFAULT 0,
    reconciled BOOL NOT NULL DEFAULT false,
    note STRING NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS transaction_category (
    id INTEGER PRIMARY KEY,
    "transaction" INTEGER NOT NULL REFERENCES "transaction" (id) ON DELETE CASCADE,
    other_account INTEGER REFERENCES account (id) ON DELETE CASCADE,
    category INTEGER REFERENCES category (id) ON DELETE CASCADE,
    outflow INTEGER NOT NULL DEFAULT 0,
    inflow INTEGER NOT NULL DEFAULT 0,
    CHECK (
        (other_account IS NOT NULL AND category IS NULL) OR
        (other_account IS NULL AND category IS NOT NULL)
    )
);

CREATE TABLE IF NOT EXISTS payee (
    id INTEGER PRIMARY KEY,
    budget INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    name STRING NOT NULL,
    UNIQUE (budget, name)
);

CREATE TABLE IF NOT EXISTS payee_default_category (
    id INTEGER PRIMARY KEY,
    payee INTEGER NOT NULL REFERENCES payee (id) ON DELETE CASCADE,
    other_account INTEGER REFERENCES account (id) ON DELETE CASCADE,
    category INTEGER REFERENCES category (id) ON DELETE CASCADE,
    percent INTEGER NOT NULL,
    CHECK (
        (other_account IS NOT NULL AND category IS NULL) OR
        (other_account IS NULL AND category IS NOT NULL)
    )
);

CREATE TABLE IF NOT EXISTS category (
    id INTEGER PRIMARY KEY,
    budget INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    name STRING NOT NULL,
    UNIQUE (budget, name)
);

CREATE TABLE IF NOT EXISTS goal (
    id INTEGER PRIMARY KEY,
    budget INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    name string NOT NULL, 
    type STRING NOT NULL,
    start DATETIME NOT NULL,
    end DATETIME,
    category INTEGER NOT NULL REFERENCES category (id) ON DELETE CASCADE,
    amount INTEGER NOT NULL,
    UNIQUE (budget, name)
);

-- +goose Down
DROP TABLE IF EXISTS goal;
DROP TABLE IF EXISTS category;
DROP TABLE IF EXISTS payee_default_category;
DROP TABLE IF EXISTS payee;
DROP TABLE IF EXISTS transaction_category;
DROP TABLE IF EXISTS "transaction";
DROP TABLE IF EXISTS account;
DROP TABLE IF EXISTS allocation;
DROP TABLE IF EXISTS budget;
