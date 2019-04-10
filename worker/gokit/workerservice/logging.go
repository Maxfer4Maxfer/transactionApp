package workerservice

import (
	"context"
	"worker/worker"

	"github.com/go-kit/kit/log"
)

// LoggingMiddleware takes a logger as a dependency
// and returns a ServiceMiddleware.
func LoggingMiddleware(logger log.Logger) func(Service) Service {
	return func(next Service) Service {
		return loggingMiddleware{logger, next}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

func (mw loggingMiddleware) Ping(ctx context.Context) (jobs int, err error) {
	defer func() {
		mw.logger.Log("method", "Ping", "jobs count", jobs, "err", err)
	}()
	return mw.next.Ping(ctx)
}

func (mw loggingMiddleware) NewJob(ctx context.Context) (id string, err error) {
	defer func() {
		mw.logger.Log("method", "NewJob", "id", id, "err", err)
	}()
	return mw.next.NewJob(ctx)
}

func (mw loggingMiddleware) GetJobs(ctx context.Context) (jobs []worker.Job, err error) {
	defer func() {
		mw.logger.Log("method", "GetJobs", "len(jobs)", len(jobs), "err", err)
	}()
	return mw.next.GetJobs(ctx)
}
