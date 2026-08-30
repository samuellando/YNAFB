-- +goose Up
CREATE TABLE reconciliation (
    account        INTEGER NOT NULL REFERENCES account (id) ON DELETE CASCADE,
    "transaction"  INTEGER NOT NULL REFERENCES "transaction" (id) ON DELETE CASCADE,
    PRIMARY KEY (account, "transaction")
);

INSERT INTO reconciliation (account, "transaction")
SELECT account, id
FROM "transaction"
WHERE reconciled;

ALTER TABLE "transaction" DROP COLUMN reconciled;

-- +goose Down
ALTER TABLE "transaction" ADD COLUMN reconciled BOOL NOT NULL DEFAULT false;

UPDATE "transaction"
SET reconciled = true
WHERE id IN (SELECT "transaction" FROM reconciliation);

DROP TABLE reconciliation;