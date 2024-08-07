// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: users.sql

package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const changeUserPassword = `-- name: ChangeUserPassword :exec
update users set updated_at=$2, password=$3 where id=$1
`

type ChangeUserPasswordParams struct {
	ID        uuid.UUID
	UpdatedAt pgtype.Timestamptz
	Password  string
}

func (q *Queries) ChangeUserPassword(ctx context.Context, arg ChangeUserPasswordParams) error {
	_, err := q.db.Exec(ctx, changeUserPassword, arg.ID, arg.UpdatedAt, arg.Password)
	return err
}

const deleteUser = `-- name: DeleteUser :exec
delete from users where id=$1
`

func (q *Queries) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteUser, id)
	return err
}

const insertUser = `-- name: InsertUser :one
insert into
    users (id, created_at, updated_at, name, mail, password)
values
    ($1, $2, $3, $4, $5, $6)
returning id, created_at, updated_at, name, mail, mail_verified_at, password
`

type InsertUserParams struct {
	ID        uuid.UUID
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	Name      string
	Mail      string
	Password  string
}

func (q *Queries) InsertUser(ctx context.Context, arg InsertUserParams) (User, error) {
	row := q.db.QueryRow(ctx, insertUser,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Name,
		arg.Mail,
		arg.Password,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Mail,
		&i.MailVerifiedAt,
		&i.Password,
	)
	return i, err
}

const queryUserByEmail = `-- name: QueryUserByEmail :one
select id, created_at, updated_at, name, mail, mail_verified_at, password from users where mail=$1
`

func (q *Queries) QueryUserByEmail(ctx context.Context, mail string) (User, error) {
	row := q.db.QueryRow(ctx, queryUserByEmail, mail)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Mail,
		&i.MailVerifiedAt,
		&i.Password,
	)
	return i, err
}

const queryUserByID = `-- name: QueryUserByID :one
select id, created_at, updated_at, name, mail, mail_verified_at, password from users where id=$1
`

func (q *Queries) QueryUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	row := q.db.QueryRow(ctx, queryUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Mail,
		&i.MailVerifiedAt,
		&i.Password,
	)
	return i, err
}

const queryUsers = `-- name: QueryUsers :many
select id, created_at, updated_at, name, mail, mail_verified_at, password from users
`

func (q *Queries) QueryUsers(ctx context.Context) ([]User, error) {
	rows, err := q.db.Query(ctx, queryUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.Mail,
			&i.MailVerifiedAt,
			&i.Password,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateUser = `-- name: UpdateUser :one
update users
    set updated_at=$2, name=$3, mail=$4, password=$5
where id = $1
returning id, created_at, updated_at, name, mail, mail_verified_at, password
`

type UpdateUserParams struct {
	ID        uuid.UUID
	UpdatedAt pgtype.Timestamptz
	Name      string
	Mail      string
	Password  string
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUser,
		arg.ID,
		arg.UpdatedAt,
		arg.Name,
		arg.Mail,
		arg.Password,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Mail,
		&i.MailVerifiedAt,
		&i.Password,
	)
	return i, err
}

const verifyUserEmail = `-- name: VerifyUserEmail :exec
update users set updated_at=$2, mail_verified_at=$3 where mail=$1
`

type VerifyUserEmailParams struct {
	Mail           string
	UpdatedAt      pgtype.Timestamptz
	MailVerifiedAt pgtype.Timestamptz
}

func (q *Queries) VerifyUserEmail(ctx context.Context, arg VerifyUserEmailParams) error {
	_, err := q.db.Exec(ctx, verifyUserEmail, arg.Mail, arg.UpdatedAt, arg.MailVerifiedAt)
	return err
}
