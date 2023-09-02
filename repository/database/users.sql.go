// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: users.sql

package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const deleteUser = `-- name: DeleteUser :exec
delete from users where id=$1
`

func (q *Queries) DeleteUser(ctx context.Context, id pgtype.UUID) error {
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
	ID        pgtype.UUID
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

const queryUser = `-- name: QueryUser :one
select id, created_at, updated_at, name, mail, mail_verified_at, password from users where id=$1
`

func (q *Queries) QueryUser(ctx context.Context, id pgtype.UUID) (User, error) {
	row := q.db.QueryRow(ctx, queryUser, id)
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
	ID        pgtype.UUID
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
