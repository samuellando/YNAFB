-- +goose Up
CREATE TABLE goal_new (
    id       INTEGER PRIMARY KEY,
    budget   INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    name     STRING  NOT NULL,
    type     STRING  NOT NULL CHECK (type IN ('monthly', 'save')),
    start    DATE    NOT NULL,
    "end"    DATE,
    category INTEGER NOT NULL REFERENCES category (id) ON DELETE CASCADE,
    amount   INTEGER NOT NULL,
    UNIQUE (budget, name),
    CHECK (type = 'monthly' OR "end" IS NOT NULL),
    CHECK ("end" IS NULL OR "end" >= start)
);

INSERT INTO goal_new (id, budget, name, type, start, "end", category, amount)
SELECT
    id,
    budget,
    name,
    CASE
        WHEN lower(type) = 'monthly' THEN 'monthly'
        WHEN lower(type) = 'save'   AND "end" IS NOT NULL THEN 'save'
        WHEN lower(type) = 'target' AND "end" IS NOT NULL THEN 'save'
        ELSE 'monthly'
    END,
    substr(start, 1, 7) || '-01',
    CASE WHEN "end" IS NOT NULL THEN substr("end", 1, 7) || '-01' END,
    category,
    amount
FROM goal;

DROP TABLE goal;
ALTER TABLE goal_new RENAME TO goal;

-- +goose Down
CREATE TABLE goal_old (
    id       INTEGER PRIMARY KEY,
    budget   INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    name     STRING  NOT NULL,
    type     STRING NOT NULL,
    start    DATETIME NOT NULL,
    "end"    DATETIME,
    category INTEGER NOT NULL REFERENCES category (id) ON DELETE CASCADE,
    amount   INTEGER NOT NULL,
    UNIQUE (budget, name)
);

INSERT INTO goal_old (id, budget, name, type, start, "end", category, amount)
SELECT id, budget, name, type, start, "end", category, amount
FROM goal;

DROP TABLE goal;
ALTER TABLE goal_old RENAME TO goal;