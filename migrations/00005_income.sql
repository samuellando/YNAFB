-- +goose Up
CREATE TABLE transaction_category_new (
    id INTEGER PRIMARY KEY,
    "transaction" INTEGER NOT NULL REFERENCES "transaction" (id) ON DELETE CASCADE,
    other_account INTEGER REFERENCES account (id) ON DELETE CASCADE,
    category INTEGER REFERENCES category (id) ON DELETE CASCADE,
    income BOOL NOT NULL DEFAULT false,
    outflow INTEGER NOT NULL DEFAULT 0,
    inflow INTEGER NOT NULL DEFAULT 0,
    CHECK (
        (CASE WHEN other_account IS NOT NULL THEN 1 ELSE 0 END
       + CASE WHEN category IS NOT NULL THEN 1 ELSE 0 END
       + CASE WHEN income THEN 1 ELSE 0 END) = 1
    )
);

INSERT INTO transaction_category_new (id, "transaction", other_account, category, income, outflow, inflow)
SELECT id, "transaction", other_account, category, false, outflow, inflow
FROM transaction_category;

DROP TABLE transaction_category;
ALTER TABLE transaction_category_new RENAME TO transaction_category;

CREATE TABLE payee_default_category_new (
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

INSERT INTO payee_default_category_new (id, payee, other_account, category, income, percent)
SELECT id, payee, other_account, category, false, percent
FROM payee_default_category;

DROP TABLE payee_default_category;
ALTER TABLE payee_default_category_new RENAME TO payee_default_category;

-- +goose Down
CREATE TABLE transaction_category_old (
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

INSERT INTO transaction_category_old (id, "transaction", other_account, category, outflow, inflow)
SELECT id, "transaction", other_account, category, outflow, inflow
FROM transaction_category
WHERE NOT income;

DROP TABLE transaction_category;
ALTER TABLE transaction_category_old RENAME TO transaction_category;

CREATE TABLE payee_default_category_old (
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

INSERT INTO payee_default_category_old (id, payee, other_account, category, percent)
SELECT id, payee, other_account, category, percent
FROM payee_default_category
WHERE NOT income;

DROP TABLE payee_default_category;
ALTER TABLE payee_default_category_old RENAME TO payee_default_category;