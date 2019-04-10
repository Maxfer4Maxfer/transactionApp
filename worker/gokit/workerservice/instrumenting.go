package workerservice

import (
	"context"
	"worker/worker"

	"github.com/go-kit/kit/metrics"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

// InstrumentingMiddleware returns a service middleware that instruments
// the number of integers summed and characters concatenated over the lifetime of
// the service.
func InstrumentingMiddleware(pings, newJobs, getJobs metrics.Counter) Middleware {
	return func(next Service) Service {
		return instrumentingMiddleware{
			pings:   pings,
			newJobs: newJobs,
			getJobs: getJobs,
			next:    next,
		}
	}
}

type instrumentingMiddleware struct {
	pings   metrics.Counter
	newJobs metrics.Counter
	getJobs metrics.Counter
	next    Service
}

func (mw instrumentingMiddleware) Ping(ctx context.Context) (int, error) {
	jobsCount, err := mw.next.Ping(ctx)
	mw.pings.Add(1)
	return jobsCount, err
}

func (mw instrumentingMiddleware) NewJob(ctx context.Context) (string, error) {
	id, err := mw.next.NewJob(ctx)
	mw.newJobs.Add(1)
	return id, err
}

func (mw instrumentingMiddleware) GetJobs(ctx context.Context) ([]worker.Job, error) {
	jobs, err := mw.next.GetJobs(ctx)
	mw.getJobs.Add(1)
	return jobs, err
}
