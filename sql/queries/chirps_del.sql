-- name: DelChirp :exec
DELETE FROM chirps
WHERE id = $1;