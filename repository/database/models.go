// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0

package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

type Queue struct {
	ID             uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ScheduledFor   time.Time
	FailedAttempts int32
	State          int32
	Message        pgtype.JSONB
	Processor      string
}

type Token struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Hash      string
	ExpiresAt time.Time
	Scope     string
	UserID    uuid.UUID
}

type User struct {
	ID             uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Name           string
	Mail           string
	MailVerifiedAt sql.NullTime
	Password       string
}
