-- +goose Up
CREATE TABLE goal_new (
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

INSERT INTO goal_new (id, budget, type, start, "end", category, amount)
SELECT g.id, g.budget, g.type, g.start, g."end", g.category, g.amount
FROM goal AS g
WHERE g.id = (
    SELECT MIN(g2.id)
    FROM goal AS g2
    WHERE g2.budget = g.budget AND g2.category = g.category
);

DROP TABLE goal;
ALTER TABLE goal_new RENAME TO goal;

-- +goose Down
CREATE TABLE goal_old (
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

INSERT INTO goal_old (id, budget, name, type, start, "end", category, amount)
SELECT id, budget, 'goal-' || id, type, start, "end", category, amount
FROM goal;

DROP TABLE goal;
ALTER TABLE goal_old RENAME TO goal;