package workerservice

import (
	"context"
	"errors"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"

	"worker/worker"
)

// Service describes a service that represents repository.
type Service interface {
	Ping(ctx context.Context) (int, error)
	NewJob(ctx context.Context) (string, error)
	GetJobs(ctx context.Context) ([]worker.Job, error)
}

// New returns a basic Service with all of the expected middlewares wired in.
func New(worker worker.Worker, logger log.Logger, pings, newJobs, getJobs metrics.Counter) Service {
	var svc Service
	{
		svc = NewBasicService(worker)
		svc = LoggingMiddleware(logger)(svc)
		svc = InstrumentingMiddleware(pings, newJobs, getJobs)(svc)
	}
	return svc
}

var (
	// ErrWorkerUnevailable allows say that something wrong happens a worker
	ErrWorkerUnevailable = errors.New("can't connect to a local repository with jobs")
)

type basicService struct {
	worker worker.Worker
}

// NewBasicService returns prepered basicService structure
func NewBasicService(worker worker.Worker) Service {
	// TODO: create pure service and put it to basicService structure
	return basicService{worker}
}

func (s basicService) Ping(_ context.Context) (int, error) {
	return s.worker.Ping()
}

func (s basicService) NewJob(_ context.Context) (string, error) {
	return s.worker.NewJob().String(), nil
}

func (s basicService) GetJobs(_ context.Context) ([]worker.Job, error) {
	return s.worker.GetJobs(), nil
}
