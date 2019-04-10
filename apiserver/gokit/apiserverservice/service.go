package apiserverservice

import (
	"context"
	"errors"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"

	"apiserver/apiserver"
	"repository/repo"
)

// Service describes a service that represents apiserver.
type Service interface {
	GetAllNodes(ctx context.Context) ([]repo.Node, error)
	NewJob(ctx context.Context) (string, error)
}

// New returns a basic Service with all of the expected middlewares wired in.
func New(api apiserver.APIServer, logger log.Logger, getAllNodes, newJobs metrics.Counter) Service {
	var svc Service
	{
		svc = NewBasicService(api)
		svc = LoggingMiddleware(logger)(svc)
		svc = InstrumentingMiddleware(getAllNodes, newJobs)(svc)
	}
	return svc
}

var (
	// ErrAPIServerUnevailable allows say that something wrong happens with a connection to DB
	ErrAPIServerUnevailable = errors.New("can't connect to a repository")
)

type basicService struct {
	apiServer apiserver.APIServer
}

// NewBasicService returns prepered basicService structure
func NewBasicService(api apiserver.APIServer) Service {
	// TODO: create pure service and put it to basicService structure
	return basicService{api}
}

func (s basicService) GetAllNodes(_ context.Context) ([]repo.Node, error) {
	return s.apiServer.GetAllNodes()
}

func (s basicService) NewJob(ctx context.Context) (string, error) {
	return s.apiServer.NewJob()
}
