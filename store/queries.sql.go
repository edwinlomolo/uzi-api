// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: queries.sql

package store

import (
	"context"
	"net/netip"

	"github.com/google/uuid"
)

const createSession = `-- name: CreateSession :one
INSERT INTO sessions (
  ip, token, user_id, expires
) VALUES (
  $1, $2, $3, $4
)
RETURNING id, ip, token, expires, user_id, created_at, updated_at
`

type CreateSessionParams struct {
	Ip      netip.Addr
	Token   string
	UserID  uuid.UUID
	Expires string
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error) {
	row := q.db.QueryRowContext(ctx, createSession,
		arg.Ip,
		arg.Token,
		arg.UserID,
		arg.Expires,
	)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.Ip,
		&i.Token,
		&i.Expires,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (
  first_name, last_name, phone
) VALUES (
  $1, $2, $3
)
RETURNING id, first_name, last_name, phone, created_at, updated_at
`

type CreateUserParams struct {
	FirstName string
	LastName  string
	Phone     string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.FirstName, arg.LastName, arg.Phone)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Phone,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const findByPhone = `-- name: FindByPhone :one
SELECT id, first_name, last_name, phone, created_at, updated_at FROM users
WHERE phone = $1
LIMIT 1
`

func (q *Queries) FindByPhone(ctx context.Context, phone string) (User, error) {
	row := q.db.QueryRowContext(ctx, findByPhone, phone)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Phone,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getSession = `-- name: GetSession :one
SELECT id, ip, token, expires, user_id, created_at, updated_at FROM
sessions
WHERE user_id = $1
`

func (q *Queries) GetSession(ctx context.Context, userID uuid.UUID) (Session, error) {
	row := q.db.QueryRowContext(ctx, getSession, userID)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.Ip,
		&i.Token,
		&i.Expires,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
