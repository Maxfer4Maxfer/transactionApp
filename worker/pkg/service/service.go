package service

import (
	"context"
	"errors"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"

	"worker/pkg/model"
)

// Service describes a service that represents repository.
type Service interface {
	Ping(ctx context.Context) (int, error)
	NewJob(ctx context.Context) (string, error)
	GetJobs(ctx context.Context) ([]model.Job, error)
}

// New returns a basic Service with all of the expected middlewares wired in.
func New(name string, IP string, port string, natsAddr string, logger log.Logger, pings, newJobs, getJobs metrics.Counter) Service {
	var svc Service
	{
		svc = NewWorker(name, IP, port, natsAddr, logger)
		svc = LoggingMiddleware(logger)(svc)
		svc = InstrumentingMiddleware(pings, newJobs, getJobs)(svc)
	}
	return svc
}

var (
	// ErrWorkerUnevailable allows say that something wrong happens a worker
	ErrWorkerUnevailable = errors.New("can't connect to a local repository with jobs")
)
