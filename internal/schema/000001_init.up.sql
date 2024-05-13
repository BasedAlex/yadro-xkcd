CREATE TABLE comics (
    id serial not null unique,
    index text not null,
    image text not null,
    keywords text array
);

CREATE TABLE indexes (
    id serial not null unique,
    data HSTORE
);