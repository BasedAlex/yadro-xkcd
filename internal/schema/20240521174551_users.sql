-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    login TEXT,
    password TEXT,
    role VARCHAR(50)
);

INSERT INTO users (login, password, role) VALUES ('alex', 'based', 'user'), ('notalex', 'notbased', 'user'), ('admin', 'admin', 'admin');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
