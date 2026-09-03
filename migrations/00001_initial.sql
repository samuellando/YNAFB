-- +goose up
CREATE TABLE budget (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE account (
    id INTEGER PRIMARY KEY,
    budget INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    UNIQUE (budget, name)
);

CREATE TABLE "transaction" (
    id INTEGER PRIMARY KEY,
    date UNIX_EPOCH_INTEGER NOT NULL,
    account INTEGER NOT NULL REFERENCES account (id) ON DELETE CASCADE,
    payee INTEGER NOT NULL REFERENCES payee (id) ON DELETE CASCADE,
    total_outflow INTEGER NOT NULL DEFAULT 0,
    total_inflow INTEGER NOT NULL DEFAULT 0,
    note TEXT NOT NULL DEFAULT ''
);

CREATE TABLE payee (
    id INTEGER PRIMARY KEY,
    budget INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    UNIQUE (budget, name)
);

CREATE TABLE "transaction_category" (
    id INTEGER PRIMARY KEY,
    "transaction" INTEGER NOT NULL REFERENCES "transaction" (id) ON DELETE CASCADE,
    other_account INTEGER REFERENCES account (id) ON DELETE CASCADE,
    category INTEGER REFERENCES category (id) ON DELETE CASCADE,
    income BOOL NOT NULL DEFAULT false,
    outflow INTEGER NOT NULL DEFAULT 0,
    inflow INTEGER NOT NULL DEFAULT 0,
    -- XOR, can only be one of category (spend), transfer, or income
    CHECK (
        (CASE WHEN other_account IS NOT NULL THEN 1 ELSE 0 END
       + CASE WHEN category IS NOT NULL THEN 1 ELSE 0 END
       + CASE WHEN income THEN 1 ELSE 0 END) = 1
    ),
    CHECK (
        NOT INCOME OR (outflow = 0 and inflow > 0)
    )
);

CREATE TABLE category (
    id INTEGER PRIMARY KEY,
    budget INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    name TEXT NOT NULL, 
    category_group INTEGER REFERENCES category_group (id) ON DELETE SET NULL,
    UNIQUE (budget, name),
    CHECK (name IS NOT 'income')
);

CREATE TABLE category_group (
    id INTEGER PRIMARY KEY,
    budget INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    UNIQUE (budget, name)
);

CREATE TABLE "payee_default_category" (
    id INTEGER PRIMARY KEY,
    payee INTEGER NOT NULL REFERENCES payee (id) ON DELETE CASCADE,
    other_account INTEGER REFERENCES account (id) ON DELETE CASCADE,
    category INTEGER REFERENCES category (id) ON DELETE CASCADE,
    income BOOL NOT NULL DEFAULT false,
    percent INTEGER NOT NULL,
    CHECK (
        (CASE WHEN other_account IS NOT NULL THEN 1 ELSE 0 END
       + CASE WHEN category IS NOT NULL THEN 1 ELSE 0 END
       + CASE WHEN income THEN 1 ELSE 0 END) = 1
    )
);

CREATE TABLE allocation (
    id INTEGER PRIMARY KEY,
    month UNIX_EPOCH_INTEGER NOT NULL,
    budget INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    category INTEGER NOT NULL REFERENCES category (id) ON DELETE CASCADE,
    amount INTEGER NOT NULL, 
    UNIQUE (budget, category, month)
);

CREATE TABLE reconciliation (
    account        INTEGER NOT NULL REFERENCES account (id) ON DELETE CASCADE,
    "transaction"  INTEGER NOT NULL REFERENCES "transaction" (id) ON DELETE CASCADE,
    PRIMARY KEY (account, "transaction")
);

CREATE TABLE "goal" (
    id       INTEGER PRIMARY KEY,
    budget   INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    type     TEXT  NOT NULL CHECK (type IN ('monthly', 'save', 'refill')),
    start    UNIX_EPOCH_INTEGER    NOT NULL,
    "end"    UNIX_EPOCH_INTEGER,
    category INTEGER NOT NULL REFERENCES category (id) ON DELETE CASCADE,
    amount   INTEGER NOT NULL,
    UNIQUE (budget, category),
    CHECK (type IN ('monthly', 'refill') OR "end" IS NOT NULL),
    CHECK ("end" IS NULL OR "end" >= start)
);
