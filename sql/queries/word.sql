-- name: CreateWord :one
INSERT INTO word (word) VALUES ($1)
RETURNING *;

-- name: GetWord :one
SELECT word FROM word
WHERE id = $1;

