package reposervice

import (
	"context"
	"errors"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"

	"repository/repo"
)

// Service describes a service that represents repository.
type Service interface {
	RegisterNode(ctx context.Context, name string, IP string, port string) (string, error)
	GetAllNodes(ctx context.Context) ([]repo.Node, error)
	NewJob(ctx context.Context) (string, error)
}

// New returns a basic Service with all of the expected middlewares wired in.
func New(repo repo.Repo, logger log.Logger, registerNodes, getAllNodes, newJobs metrics.Counter) Service {
	var svc Service
	{
		svc = NewBasicService(repo)
		svc = LoggingMiddleware(logger)(svc)
		svc = InstrumentingMiddleware(registerNodes, getAllNodes, newJobs)(svc)
	}
	return svc
}

var (
	// ErrRepoUnevailable allows say that something wrong happens with a connection to DB
	ErrRepoUnevailable = errors.New("can't connect to a storage service")

	// ErrNodeAlreadyExist prevents users add a node with dublicate name
	ErrNodeAlreadyExist = errors.New("node with same name already registered in repo")

	// ErrEmptyRepo shows that a repo is empty
	ErrEmptyRepo = errors.New("empty repository")
)

type basicService struct {
	repo repo.Repo
}

// NewBasicService returns prepered basicService structure
func NewBasicService(repo repo.Repo) Service {
	// TODO: create pure service and put it to basicService structure
	return basicService{repo}
}

func (s basicService) RegisterNode(_ context.Context, name string, IP string, port string) (string, error) {
	return s.repo.RegisterNode(name, IP, port)
}

func (s basicService) GetAllNodes(_ context.Context) ([]repo.Node, error) {
	return s.repo.GetAllNodes()
}

func (s basicService) NewJob(ctx context.Context) (string, error) {
	return s.repo.NewJob()
}
