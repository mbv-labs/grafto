// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0

package database

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID             pgtype.UUID
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	Name           string
	Mail           string
	MailVerifiedAt pgtype.Timestamptz
	Password       string
}
