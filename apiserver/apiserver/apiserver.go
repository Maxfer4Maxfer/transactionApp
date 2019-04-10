package apiserver

import (
	"context"
	"errors"
	"time"

	"repository/gokit/repotransport"
	"repository/repo"

	stdopentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/log"
)

// APIServer implements APIServer interface
type APIServer struct {
	IP     string
	Port   string
	logger log.Logger
}

// New create new APIServer
func New(IP string, port string, logger log.Logger) APIServer {
	api := APIServer{IP, port, logger}
	return api
}

var (
	// ErrAPIServerUnevailable allows say that something wrong happens with a connection to DB
	ErrAPIServerUnevailable = errors.New("can't connect to a repository")
)

// GetAllNodes returs all available nodes with their jobs
func (api *APIServer) GetAllNodes() ([]repo.Node, error) {
	grpcAddr := api.IP + api.Port
	api.logger.Log("method", "GetAllNodes", "connecting to ", grpcAddr)

	ctx, close := context.WithTimeout(context.Background(), time.Second)
	defer close()
	conn, err := grpc.DialContext(ctx, grpcAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		api.logger.Log("method", "GetAllNodes", "err", err)
	}
	defer conn.Close()

	otTracer := stdopentracing.GlobalTracer() // no-op
	svc := repotransport.NewGRPCClient(conn, otTracer, api.logger)

	nodes, err := svc.GetAllNodes(context.Background())
	api.logger.Log("method", "GetAllNodes", "nodes", nodes)

	return nodes, err
}

// NewJob starts new job on a free node
func (api *APIServer) NewJob() (string, error) {

	grpcAddr := api.IP + api.Port
	api.logger.Log("method", "NewJob", "connecting to ", grpcAddr)

	ctx, close := context.WithTimeout(context.Background(), time.Second)
	defer close()
	conn, err := grpc.DialContext(ctx, grpcAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		api.logger.Log("method", "NewJob", "err", err)
	}
	defer conn.Close()

	otTracer := stdopentracing.GlobalTracer() // no-op
	svc := repotransport.NewGRPCClient(conn, otTracer, api.logger)

	jID, err := svc.NewJob(context.Background())
	api.logger.Log("method", "NewJob", "job ID", jID)

	return jID, err
}
