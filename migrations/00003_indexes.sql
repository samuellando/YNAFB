-- +goose up
CREATE INDEX trx_date ON trx(budget_id, date, id);
CREATE INDEX trx_account ON trx(account_id);
CREATE INDEX trx_payee ON trx(payee_id);
CREATE INDEX trx_line_trx ON trx_line(trx_id);
CREATE INDEX trx_line_category ON trx_line(category_id);
CREATE INDEX trx_line_dest_account ON trx_line(dest_account_id);
CREATE INDEX category_category_group ON category(category_group_id);
CREATE INDEX reconciliation_account ON reconciliation(account_id);
CREATE INDEX reconciliation_trx_account ON reconciliation(trx_id, account_id);

-- +goose down
DROP INDEX trx_account;
DROP INDEX trx_payee;
DROP INDEX trx_line_trx;
DROP INDEX trx_line_category;
DROP INDEX trx_line_dest_account;
DROP INDEX category_category_group;
DROP INDEX reconciliation_account;
DROP INDEX reconciliation_trx;
