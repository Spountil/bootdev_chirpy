-- name: GetAllChrips :many
SELECT *
FROM chirps
ORDER BY created_at;
