-- name: CreateChirp :one
INSERT INTO chirps(body, user_id)
VALUES ($1,$2)
RETURNING *;

-- name: GetAllChirps :many
SELECT * FROM chirps
ORDER BY
    CASE WHEN COALESCE($1, 'asc') = 'desc' THEN created_at END DESC,
    CASE WHEN COALESCE($1, 'asc') = 'asc' THEN created_at END ASC;

-- name: GetChirpByID :one
SELECT * FROM chirps
WHERE id = $1;

-- name: DeleteChirpById :exec
DELETE FROM chirps
WHERE id = $1;

-- name: GetChirpsByAuthorID :many
SELECT * FROM chirps
WHERE user_id = $1
ORDER BY
    CASE WHEN COALESCE($2, 'asc') = 'desc' THEN created_at END DESC,
    CASE WHEN COALESCE($2, 'asc') = 'asc' THEN created_at END ASC;
