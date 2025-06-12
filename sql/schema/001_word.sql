-- +goose Up
CREATE TABLE word (
    id      serial     PRIMARY KEY,
    word    varchar(255) NOT NULL unique,
    partOfSpeech    varchar(64) DEFAULT 'unknown'
);


-- +goose Down
DROP TABLE IF EXISTS word;
