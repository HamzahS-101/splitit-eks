-- +goose Up
CREATE INDEX expenses_paid_by_idx ON expenses (paid_by);
CREATE INDEX expense_shares_member_idx ON expense_shares (member_id);
CREATE INDEX expense_shares_group_idx ON expense_shares (group_id);
CREATE INDEX settlements_from_idx ON settlements (from_member);
CREATE INDEX settlements_to_idx ON settlements (to_member);

-- +goose Down
DROP INDEX settlements_to_idx;
DROP INDEX settlements_from_idx;
DROP INDEX expense_shares_group_idx;
DROP INDEX expense_shares_member_idx;
DROP INDEX expenses_paid_by_idx;
