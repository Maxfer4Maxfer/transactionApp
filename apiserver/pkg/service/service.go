package service

import (
	"context"
	"errors"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	stdopentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"

	repo "repository/pkg/model"
	repotransport "repository/pkg/transport"

)

// Service describes a service that represents apiserver.
type Service interface {
	GetAllNodes(ctx context.Context) ([]repo.Node, error)
	NewJob(ctx context.Context) (string, error)
}


// New returns a basic Service with all of the expected middlewares wired in.
func New(IP string, port string, logger log.Logger, getAllNodes, newJobs metrics.Counter) Service {
	var svc Service
	{
		svc = APIServer{IP, port, logger}
		svc = LoggingMiddleware(logger)(svc)
		svc = InstrumentingMiddleware(getAllNodes, newJobs)(svc)
	}
	return svc
}

var (
	// ErrAPIServerUnevailable allows say that something wrong happens with a connection to DB
	ErrAPIServerUnevailable = errors.New("can't connect to a repository")
)


// APIServer implements Service interface
type APIServer struct {
	IP     string
	Port   string
	logger log.Logger
}

// GetAllNodes returs all available nodes with their jobs
func (api APIServer) GetAllNodes(ctx context.Context) ([]repo.Node, error) {
	grpcAddr := api.IP + api.Port
	api.logger.Log("method", "GetAllNodes", "connecting to ", grpcAddr)

	ctx, close := context.WithTimeout(ctx, time.Second)
	defer close()
	conn, err := grpc.DialContext(ctx, grpcAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		api.logger.Log("method", "GetAllNodes", "err", err)
	}
	defer conn.Close()

	otTracer := stdopentracing.GlobalTracer() // no-op
	svc := repotransport.NewGRPCClient(conn, otTracer, api.logger)

	nodes, err := svc.GetAllNodes(ctx)
	api.logger.Log("method", "GetAllNodes", "nodes", nodes)

	return nodes, err
}

// NewJob starts new job on a free node
func (api APIServer) NewJob(ctx context.Context) (string, error) {
	grpcAddr := api.IP + api.Port
	api.logger.Log("method", "NewJob", "connecting to ", grpcAddr)

	ctx, close := context.WithTimeout(ctx, time.Second)
	defer close()
	conn, err := grpc.DialContext(ctx, grpcAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		api.logger.Log("method", "NewJob", "err", err)
	}
	defer conn.Close()

	otTracer := stdopentracing.GlobalTracer() // no-op
	svc := repotransport.NewGRPCClient(conn, otTracer, api.logger)

	jID, err := svc.NewJob(ctx)
	api.logger.Log("method", "NewJob", "job ID", jID)

	return jID, err
}
