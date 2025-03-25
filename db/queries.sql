-- name: CreateUser :one
INSERT INTO users (google_id, email, name, picture)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, expires_at)
VALUES ($1, $2, $3);

-- name: GetUserByGoogleID :one
SELECT * FROM users
WHERE google_id = $1
LIMIT 1;

-- name: GetSessionWithUser :one
SELECT s.*, u.*
FROM sessions s
INNER JOIN users u ON s.user_id = u.id
WHERE s.id = $1
LIMIT 1;

-- name: UpdateSessionExpiration :exec
UPDATE sessions
SET expires_at = $1
WHERE id = $2;

-- name: DeleteSessionByID :exec
DELETE FROM sessions
WHERE id = $1;
