-- +goose Up
-- +goose StatementBegin
CREATE TABLE comics (
    id serial not null unique,
    index text not null,
    image text not null,
    keywords text array,
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now()
);

CREATE TABLE indexes (
    stem text not null,
    comics integer array not null
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE comics;

DROP TABLE indexes;
-- +goose StatementEnd
