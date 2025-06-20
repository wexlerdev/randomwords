// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: word.sql

package database

import (
	"context"
)

const createWord = `-- name: CreateWord :one
INSERT INTO word (word) VALUES ($1)
RETURNING id, word, partofspeech
`

func (q *Queries) CreateWord(ctx context.Context, word string) (Word, error) {
	row := q.db.QueryRowContext(ctx, createWord, word)
	var i Word
	err := row.Scan(&i.ID, &i.Word, &i.Partofspeech)
	return i, err
}

const getWord = `-- name: GetWord :one
SELECT word FROM word
WHERE id = $1
`

func (q *Queries) GetWord(ctx context.Context, id int32) (string, error) {
	row := q.db.QueryRowContext(ctx, getWord, id)
	var word string
	err := row.Scan(&word)
	return word, err
}
