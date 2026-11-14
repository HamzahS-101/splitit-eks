-- +goose Up
CREATE TABLE groups (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug             text NOT NULL UNIQUE,
    name             text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 80),
    currency         text NOT NULL DEFAULT 'GBP',
    created_at       timestamptz NOT NULL DEFAULT now(),
    last_activity_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX groups_last_activity_idx ON groups (last_activity_at);

CREATE TABLE members (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id   uuid NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    name       text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 40),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (group_id, name),
    UNIQUE (group_id, id)
);

CREATE TABLE expenses (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id     uuid NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    paid_by      uuid NOT NULL,
    description  text NOT NULL CHECK (char_length(description) BETWEEN 1 AND 120),
    amount_pence bigint NOT NULL CHECK (amount_pence > 0),
    split_type   text NOT NULL CHECK (split_type IN ('equal', 'custom')),
    spent_on     date NOT NULL DEFAULT current_date,
    created_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (group_id, id),
    FOREIGN KEY (group_id, paid_by) REFERENCES members (group_id, id)
);

CREATE INDEX expenses_group_idx ON expenses (group_id, spent_on DESC, created_at DESC);

CREATE TABLE expense_shares (
    expense_id   uuid NOT NULL,
    group_id     uuid NOT NULL,
    member_id    uuid NOT NULL,
    amount_pence bigint NOT NULL CHECK (amount_pence >= 0),
    PRIMARY KEY (expense_id, member_id),
    FOREIGN KEY (group_id, expense_id) REFERENCES expenses (group_id, id) ON DELETE CASCADE,
    FOREIGN KEY (group_id, member_id) REFERENCES members (group_id, id)
);

CREATE TABLE settlements (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id     uuid NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    from_member  uuid NOT NULL,
    to_member    uuid NOT NULL,
    amount_pence bigint NOT NULL CHECK (amount_pence > 0),
    created_at   timestamptz NOT NULL DEFAULT now(),
    CHECK (from_member <> to_member),
    FOREIGN KEY (group_id, from_member) REFERENCES members (group_id, id),
    FOREIGN KEY (group_id, to_member) REFERENCES members (group_id, id)
);

CREATE INDEX settlements_group_idx ON settlements (group_id);

-- +goose Down
DROP TABLE settlements;
DROP TABLE expense_shares;
DROP TABLE expenses;
DROP TABLE members;
DROP TABLE groups;
