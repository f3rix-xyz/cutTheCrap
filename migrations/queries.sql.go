// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: queries.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createSession = `-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, expires_at)
VALUES ($1, $2, $3)
`

type CreateSessionParams struct {
	ID        string
	UserID    int32
	ExpiresAt pgtype.Timestamptz
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) error {
	_, err := q.db.Exec(ctx, createSession, arg.ID, arg.UserID, arg.ExpiresAt)
	return err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (google_id, email, name, picture)
VALUES ($1, $2, $3, $4)
RETURNING id, google_id, email, name, picture
`

type CreateUserParams struct {
	GoogleID string
	Email    string
	Name     string
	Picture  string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.GoogleID,
		arg.Email,
		arg.Name,
		arg.Picture,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.GoogleID,
		&i.Email,
		&i.Name,
		&i.Picture,
	)
	return i, err
}

const deleteSessionByID = `-- name: DeleteSessionByID :exec
DELETE FROM sessions
WHERE id = $1
`

func (q *Queries) DeleteSessionByID(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteSessionByID, id)
	return err
}

const getSessionWithUser = `-- name: GetSessionWithUser :one
SELECT s.id, s.user_id, s.expires_at, u.id, u.google_id, u.email, u.name, u.picture
FROM sessions s
INNER JOIN users u ON s.user_id = u.id
WHERE s.id = $1
LIMIT 1
`

type GetSessionWithUserRow struct {
	ID        string
	UserID    int32
	ExpiresAt pgtype.Timestamptz
	ID_2      int32
	GoogleID  string
	Email     string
	Name      string
	Picture   string
}

func (q *Queries) GetSessionWithUser(ctx context.Context, id string) (GetSessionWithUserRow, error) {
	row := q.db.QueryRow(ctx, getSessionWithUser, id)
	var i GetSessionWithUserRow
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ExpiresAt,
		&i.ID_2,
		&i.GoogleID,
		&i.Email,
		&i.Name,
		&i.Picture,
	)
	return i, err
}

const getUserByGoogleID = `-- name: GetUserByGoogleID :one
SELECT id, google_id, email, name, picture FROM users
WHERE google_id = $1
LIMIT 1
`

func (q *Queries) GetUserByGoogleID(ctx context.Context, googleID string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByGoogleID, googleID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.GoogleID,
		&i.Email,
		&i.Name,
		&i.Picture,
	)
	return i, err
}

const updateSessionExpiration = `-- name: UpdateSessionExpiration :exec
UPDATE sessions
SET expires_at = $1
WHERE id = $2
`

type UpdateSessionExpirationParams struct {
	ExpiresAt pgtype.Timestamptz
	ID        string
}

func (q *Queries) UpdateSessionExpiration(ctx context.Context, arg UpdateSessionExpirationParams) error {
	_, err := q.db.Exec(ctx, updateSessionExpiration, arg.ExpiresAt, arg.ID)
	return err
}
