// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: queue.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

const clearQueue = `-- name: ClearQueue :exec
delete from queue
`

func (q *Queries) ClearQueue(ctx context.Context) error {
	_, err := q.db.Exec(ctx, clearQueue)
	return err
}

const deleteJob = `-- name: DeleteJob :exec
delete from queue where id = $1
`

func (q *Queries) DeleteJob(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteJob, id)
	return err
}

const failJob = `-- name: FailJob :exec
update queue
    SET state = $1, updated_at = $2, scheduled_for = $3, failed_attempts = failed_attempts + 1
WHERE id = $4
`

type FailJobParams struct {
	State        int32
	UpdatedAt    time.Time
	ScheduledFor time.Time
	ID           uuid.UUID
}

func (q *Queries) FailJob(ctx context.Context, arg FailJobParams) error {
	_, err := q.db.Exec(ctx, failJob,
		arg.State,
		arg.UpdatedAt,
		arg.ScheduledFor,
		arg.ID,
	)
	return err
}

const insertJob = `-- name: InsertJob :exec
insert into queue
    (id, created_at, updated_at, scheduled_for, failed_attempts, state, message, processor)
values
    ($1, $2, $3, $4, $5, $6, $7, $8)
`

type InsertJobParams struct {
	ID             uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ScheduledFor   time.Time
	FailedAttempts int32
	State          int32
	Message        pgtype.JSONB
	Processor      string
}

func (q *Queries) InsertJob(ctx context.Context, arg InsertJobParams) error {
	_, err := q.db.Exec(ctx, insertJob,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.ScheduledFor,
		arg.FailedAttempts,
		arg.State,
		arg.Message,
		arg.Processor,
	)
	return err
}

const queryJobs = `-- name: QueryJobs :many
update queue
    set state = $1, updated_at = $2
    where id in (
        select id
        from queue as inner_queue
        where inner_queue.state = $4::int 
        and inner_queue.scheduled_for::time <= $5::time 
        and inner_queue.failed_attempts < $6::int
        order by inner_queue.scheduled_for
        for update skip locked
        limit $3
    )
returning id, created_at, updated_at, scheduled_for, failed_attempts, state, message, processor
`

type QueryJobsParams struct {
	State               int32
	UpdatedAt           time.Time
	Limit               int32
	InnerState          int32
	InnerScheduledFor   time.Time
	InnerFailedAttempts int32
}

func (q *Queries) QueryJobs(ctx context.Context, arg QueryJobsParams) ([]Queue, error) {
	rows, err := q.db.Query(ctx, queryJobs,
		arg.State,
		arg.UpdatedAt,
		arg.Limit,
		arg.InnerState,
		arg.InnerScheduledFor,
		arg.InnerFailedAttempts,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Queue
	for rows.Next() {
		var i Queue
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.ScheduledFor,
			&i.FailedAttempts,
			&i.State,
			&i.Message,
			&i.Processor,
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
