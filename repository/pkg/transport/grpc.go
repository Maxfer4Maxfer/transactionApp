package transport

import (
	"context"
	"errors"
	"time"

	"github.com/go-kit/kit/ratelimit"
	"google.golang.org/grpc"

	stdopentracing "github.com/opentracing/opentracing-go"

	timestamp "github.com/golang/protobuf/ptypes"

	"github.com/go-kit/kit/circuitbreaker"
	kitendpoint "github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/google/uuid"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"

	grpctransport "github.com/go-kit/kit/transport/grpc"

	"repository/pkg/endpoint"
	"repository/pkg/service"
	pb "repository/pb"
	repo "repository/pkg/model"
)

type grpcServer struct {
	registerNode grpctransport.Handler
	getAllNodes  grpctransport.Handler
	newJob       grpctransport.Handler
}

// NewGRPCServer makes a set of endpoints available as a gRPC AddServer.
func NewGRPCServer(endpoints endpoint.EndpointSet, otTracer stdopentracing.Tracer, logger log.Logger) pb.RepoServer {

	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorLogger(logger),
	}
	return &grpcServer{
		registerNode: grpctransport.NewServer(
			endpoints.RegisterNodeEndpoint,
			decodeGRPCRegisterNodeRequest,
			encodeGRPCRegisterNodeResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "RegisterNode", logger)))...,
		),
		getAllNodes: grpctransport.NewServer(
			endpoints.GetAllNodesEndpoint,
			decodeGRPCGetAllNodesRequest,
			encodeGRPCGetAllNodesResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "GetAllNodes", logger)))...,
		),
		newJob: grpctransport.NewServer(
			endpoints.NewJobEndpoint,
			decodeGRPCNewJobRequest,
			encodeGRPCNewJobResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "NewJob", logger)))...,
		),
	}
}

func (s *grpcServer) RegisterNode(ctx context.Context, req *pb.RegisterNodeRequest) (*pb.RegisterNodeReply, error) {
	_, rep, err := s.registerNode.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.RegisterNodeReply), nil
}

func (s *grpcServer) GetAllNodes(ctx context.Context, req *pb.GetAllNodesRequest) (*pb.GetAllNodesReply, error) {
	_, rep, err := s.getAllNodes.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.GetAllNodesReply), nil
}

func (s *grpcServer) NewJob(ctx context.Context, req *pb.NewJobRequest) (*pb.NewJobReply, error) {
	_, rep, err := s.newJob.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.NewJobReply), nil
}

// NewGRPCClient returns an RepoService backed by a gRPC server at the other end
// of the conn. The caller is responsible for constructing the conn, and
// eventually closing the underlying transport. We bake-in certain middlewares,
// implementing the client library pattern.
func NewGRPCClient(conn *grpc.ClientConn, otTracer stdopentracing.Tracer, logger log.Logger) service.Service {
	// We construct a single ratelimiter middleware, to limit the total outgoing
	// QPS from this client to all methods on the remote instance. We also
	// construct per-endpoint circuitbreaker middlewares to demonstrate how
	// that's done, although they could easily be combined into a single breaker
	// for the entire remote instance, too.
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))

	// global client middlewares
	options := []grpctransport.ClientOption{}

	// Each individual endpoint is an grpc/transport.Client (which implements
	// kitendpoint.Endpoint) that gets wrapped with various middlewares. If you
	// made your own client library, you'd do this work there, so your server
	// could rely on a consistent set of client behavior.
	var registerNodeEndpoint kitendpoint.Endpoint
	{
		registerNodeEndpoint = grpctransport.NewClient(
			conn,
			"pb.repo.Repo",
			"RegisterNode",
			encodeGRPCRegisterNodeRequest,
			decodeGRPCRegisterNodeResponse,
			pb.RegisterNodeReply{},
			append(options, grpctransport.ClientBefore(opentracing.ContextToGRPC(otTracer, logger)))...,
		).Endpoint()
		registerNodeEndpoint = opentracing.TraceClient(otTracer, "RegisterNode")(registerNodeEndpoint)
		registerNodeEndpoint = limiter(registerNodeEndpoint)
		registerNodeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "RegisterNode",
			Timeout: 30 * time.Second,
		}))(registerNodeEndpoint)
	}

	var getAllNodesEndpoint kitendpoint.Endpoint
	{
		getAllNodesEndpoint = grpctransport.NewClient(
			conn,
			"pb.repo.Repo",
			"GetAllNodes",
			encodeGRPCGetAllNodesRequest,
			decodeGRPCGetAllNodesResponse,
			pb.GetAllNodesReply{},
			append(options, grpctransport.ClientBefore(opentracing.ContextToGRPC(otTracer, logger)))...,
		).Endpoint()
		getAllNodesEndpoint = opentracing.TraceClient(otTracer, "GetAllNodes")(getAllNodesEndpoint)
		getAllNodesEndpoint = limiter(getAllNodesEndpoint)
		getAllNodesEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "GetAllNodes",
			Timeout: 30 * time.Second,
		}))(getAllNodesEndpoint)
	}

	var newJobEndpoint kitendpoint.Endpoint
	{
		newJobEndpoint = grpctransport.NewClient(
			conn,
			"pb.repo.Repo",
			"NewJob",
			encodeGRPCNewJobRequest,
			decodeGRPCNewJobResponse,
			pb.NewJobReply{},
			append(options, grpctransport.ClientBefore(opentracing.ContextToGRPC(otTracer, logger)))...,
		).Endpoint()
		newJobEndpoint = opentracing.TraceClient(otTracer, "NewJob")(newJobEndpoint)
		newJobEndpoint = limiter(newJobEndpoint)
		newJobEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "NewJob",
			Timeout: 30 * time.Second,
		}))(newJobEndpoint)
	}

	// Returning the endpoint.EndpointSet as a service.Service relies on the
	// endpoint.EndpointSet implementing the Service methods. That's just a simple bit
	// of glue code.
	return endpoint.EndpointSet{
		RegisterNodeEndpoint: registerNodeEndpoint,
		GetAllNodesEndpoint:  getAllNodesEndpoint,
		NewJobEndpoint:       newJobEndpoint,
	}
}

// ********** Common **********

// These annoying helper functions are required to translate Go error types to
// and from strings, which is the type we use in our IDLs to represent errors.
// There is special casing to treat empty strings as nil errors.

func str2err(s string) error {
	if s == "" {
		return nil
	}
	return errors.New(s)
}

func err2str(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// ********** RegisterNode **********

// encodeGRPCRegisterNodeRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain RegisterNode request to a gRPC RegisterNode request. Primarily useful in a client.
func encodeGRPCRegisterNodeRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(endpoint.RegisterNodeRequest)
	return &pb.RegisterNodeRequest{Name: req.Name, NodeIP: req.IP, NodePort: req.Port}, nil
}

// decodeGRPCRegisterNodeRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC RegisterNode request to a user-domain RegisterNode request. Primarily useful in a server.
func decodeGRPCRegisterNodeRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.RegisterNodeRequest)
	return endpoint.RegisterNodeRequest{Name: req.Name, IP: req.NodeIP, Port: req.NodePort}, nil
}

// encodeGRPCRegisterNodeResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain RegisterNode response to a gRPC RegisterNode reply. Primarily useful in a server.
func encodeGRPCRegisterNodeResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(endpoint.RegisterNodeResponse)
	return &pb.RegisterNodeReply{NodeID: resp.ID, Err: err2str(resp.Err)}, nil
}

// decodeGRPCRegisterNodeResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC RegisterNode reply to a user-domain RegisterNode response. Primarily useful in a client.
func decodeGRPCRegisterNodeResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.RegisterNodeReply)
	return endpoint.RegisterNodeResponse{ID: reply.NodeID, Err: str2err(reply.Err)}, nil
}

// ********** GetAllNodes **********

// encodeGRPCGetAllNodesRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain GetAllNodes request to a gRPC GetAllNodes request. Primarily useful in a client.
func encodeGRPCGetAllNodesRequest(_ context.Context, request interface{}) (interface{}, error) {
	// req := request.(endpoint.GetAllNodesRequest)
	return &pb.GetAllNodesRequest{}, nil
}

// decodeGRPCGetAllNodesRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC GetAllNodes request to a user-domain GetAllNodes request. Primarily useful in a server.
func decodeGRPCGetAllNodesRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	// req := grpcReq.(*pb.GetAllNodesRequest)
	return endpoint.GetAllNodesRequest{}, nil
}

// encodeGRPCGetAllNodesResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain GetAllNodes response to a gRPC GetAllNodes reply. Primarily useful in a server.
func encodeGRPCGetAllNodesResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(endpoint.GetAllNodesResponse)

	// conver []repo.Node to []*pb.Node
	pbNodes := make([]*pb.Node, 0)
	for _, n := range resp.Nodes {

		// conver []repo.Job to []*pb.Job
		pbJobs := make([]*pb.Job, 0)
		for _, j := range n.Jobs {
			st, _ := timestamp.TimestampProto(j.StartTime)
			ft, _ := timestamp.TimestampProto(j.FinishTime)
			pbJob := &pb.Job{
				ID:         j.ID.String(),
				Per:        j.Per,
				Duration:   j.Duration,
				StartTime:  st,
				FinishTime: ft,
			}
			pbJobs = append(pbJobs, pbJob)
		}

		pbNode := &pb.Node{
			ID:        n.ID.String(),
			Name:      n.Name,
			IP:        n.IP,
			Port:      n.Port,
			JobsCount: int32(n.JobsCount),
			Jobs:      pbJobs,
		}
		pbNodes = append(pbNodes, pbNode)
	}

	return &pb.GetAllNodesReply{Nodes: pbNodes, Err: err2str(resp.Err)}, nil
}

// decodeGRPCGetAllNodesResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC GetAllNodes reply to a user-domain GetAllNodes response. Primarily useful in a client.
func decodeGRPCGetAllNodesResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.GetAllNodesReply)

	// conver []*pb.Node to []repo.Node
	nodes := make([]repo.Node, 0)
	for _, n := range reply.Nodes {
		// conver []*pb.Job to []repo.Job
		jobs := make([]repo.Job, 0)
		for _, j := range n.Jobs {
			id, _ := uuid.Parse(j.ID)
			st, _ := timestamp.Timestamp(j.StartTime)
			ft, _ := timestamp.Timestamp(j.FinishTime)
			job := repo.Job{
				ID:         repo.JobID{id},
				Per:        j.Per,
				Duration:   j.Duration,
				StartTime:  st,
				FinishTime: ft,
			}
			jobs = append(jobs, job)
		}

		id, _ := uuid.Parse(n.ID)
		node := repo.Node{
			ID:        repo.NodeID{id},
			Name:      n.Name,
			IP:        n.IP,
			Port:      n.Port,
			JobsCount: int(n.JobsCount),
			Jobs:      jobs,
		}
		nodes = append(nodes, node)
	}

	return endpoint.GetAllNodesResponse{Nodes: nodes, Err: str2err(reply.Err)}, nil
}

// ********** NewJob **********

// encodeGRPCNewJobRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain NewJob request to a gRPC NewJob request. Primarily useful in a client.
func encodeGRPCNewJobRequest(_ context.Context, request interface{}) (interface{}, error) {
	// req := request.(endpoint.NewJobRequest)
	return &pb.NewJobRequest{}, nil
}

// decodeGRPCNewJobRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC NewJob request to a user-domain NewJob request. Primarily useful in a server.
func decodeGRPCNewJobRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	// req := grpcReq.(*pb.NewJobRequest)
	return endpoint.NewJobRequest{}, nil
}

// encodeGRPCNewJobResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain NewJob response to a gRPC NewJob reply. Primarily useful in a server.
func encodeGRPCNewJobResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(endpoint.NewJobResponse)
	return &pb.NewJobReply{ID: resp.ID, Err: err2str(resp.Err)}, nil
}

// decodeGRPCNewJobResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC NewJob reply to a user-domain NewJob response. Primarily useful in a client.
func decodeGRPCNewJobResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.NewJobReply)
	return endpoint.NewJobResponse{ID: reply.ID, Err: str2err(reply.Err)}, nil
}
