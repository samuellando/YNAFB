-- +goose Up
CREATE TABLE goal_new (
    id       INTEGER PRIMARY KEY,
    budget   INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    type     STRING  NOT NULL CHECK (type IN ('monthly', 'save', 'refill')),
    start    DATE    NOT NULL,
    "end"    DATE,
    category INTEGER NOT NULL REFERENCES category (id) ON DELETE CASCADE,
    amount   INTEGER NOT NULL,
    UNIQUE (budget, category),
    CHECK (type IN ('monthly', 'refill') OR "end" IS NOT NULL),
    CHECK ("end" IS NULL OR "end" >= start)
);

INSERT INTO goal_new (id, budget, type, start, "end", category, amount)
SELECT id, budget, type, start, "end", category, amount
FROM goal;

DROP TABLE goal;
ALTER TABLE goal_new RENAME TO goal;

-- +goose Down
CREATE TABLE goal_old (
    id       INTEGER PRIMARY KEY,
    budget   INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    type     STRING  NOT NULL CHECK (type IN ('monthly', 'save')),
    start    DATE    NOT NULL,
    "end"    DATE,
    category INTEGER NOT NULL REFERENCES category (id) ON DELETE CASCADE,
    amount   INTEGER NOT NULL,
    UNIQUE (budget, category),
    CHECK (type = 'monthly' OR "end" IS NOT NULL),
    CHECK ("end" IS NULL OR "end" >= start)
);

INSERT INTO goal_old (id, budget, type, start, "end", category, amount)
SELECT id, budget, type, start, "end", category, amount
FROM goal;

DROP TABLE goal;
ALTER TABLE goal_old RENAME TO goal;