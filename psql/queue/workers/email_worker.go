package workers

import (
	"context"

	"github.com/mbv-labs/grafto/psql/queue/jobs"
	"github.com/mbv-labs/grafto/services"
	"github.com/riverqueue/river"
)

type EmailJobWorker struct {
	emailer services.EmailClient
	river.WorkerDefaults[jobs.EmailJobArgs]
}

func (w *EmailJobWorker) Work(ctx context.Context, job *river.Job[jobs.EmailJobArgs]) error {
	return w.emailer.SendEmail(
		ctx,
		services.EmailPayload{
			To:       job.Args.To,
			From:     job.Args.From,
			Subject:  job.Args.Subject,
			HtmlBody: job.Args.TextVersion,
			TextBody: job.Args.HtmlVersion,
		},
	)
}
