-- +goose Up
CREATE TABLE IF NOT EXISTS category_group (
    id INTEGER PRIMARY KEY,
    budget INTEGER NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    name STRING NOT NULL,
    UNIQUE (budget, name)
);

ALTER TABLE "category" 
ADD COLUMN category_group INTEGER REFERENCES category_group (id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE "category" DROP COLUMN category_group;

DROP TABLE category_group;
