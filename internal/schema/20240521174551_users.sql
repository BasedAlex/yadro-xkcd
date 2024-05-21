-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    login TEXT,
    password TEXT
);

INSERT INTO users (login, password) VALUES ('alex', 'based');
INSERT INTO users (login, password) VALUES ('notalex', 'notbased');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
