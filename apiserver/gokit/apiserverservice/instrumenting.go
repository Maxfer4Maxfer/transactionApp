package apiserverservice

import (
	"context"
	"repository/repo"

	"github.com/go-kit/kit/metrics"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

// InstrumentingMiddleware returns a service middleware that instruments
// the number of integers summed and characters concatenated over the lifetime of
// the service.
func InstrumentingMiddleware(getAllNodes, newJobs metrics.Counter) Middleware {
	return func(next Service) Service {
		return instrumentingMiddleware{
			getAllNodes: getAllNodes,
			newJobs:     newJobs,
			next:        next,
		}
	}
}

type instrumentingMiddleware struct {
	getAllNodes metrics.Counter
	newJobs     metrics.Counter
	next        Service
}

func (mw instrumentingMiddleware) GetAllNodes(ctx context.Context) ([]repo.Node, error) {
	nodes, err := mw.next.GetAllNodes(ctx)
	mw.getAllNodes.Add(1)
	return nodes, err
}

func (mw instrumentingMiddleware) NewJob(ctx context.Context) (string, error) {
	id, err := mw.next.NewJob(ctx)
	mw.newJobs.Add(1)
	return id, err
}
