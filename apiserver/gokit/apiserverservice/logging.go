package apiserverservice

import (
	"context"
	"repository/repo"

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

func (mw loggingMiddleware) GetAllNodes(ctx context.Context) (nodes []repo.Node, err error) {
	defer func() {
		mw.logger.Log("method", "getAllNodes", "len(nodes)", len(nodes), "err", err)
	}()
	return mw.next.GetAllNodes(ctx)
}

func (mw loggingMiddleware) NewJob(ctx context.Context) (ID string, err error) {
	defer func() {
		mw.logger.Log("method", "newJob", "id", ID, "err", err)
	}()
	return mw.next.NewJob(ctx)
}
