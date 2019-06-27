package transport

import (
	"context"
	"errors"
	"fmt"
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

	"worker/pkg/endpoint"
	"worker/pkg/service"
	worker "worker/pkg/model"
	pb "worker/pb"
)

type grpcServer struct {
	ping    grpctransport.Handler
	getJobs grpctransport.Handler
	newJob  grpctransport.Handler
}

// NewGRPCServer makes a set of endpoints available as a gRPC AddServer.
func NewGRPCServer(endpoints endpoint.EndpointSet, otTracer stdopentracing.Tracer, logger log.Logger) pb.WorkerServer {

	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorLogger(logger),
	}

	return &grpcServer{
		ping: grpctransport.NewServer(
			endpoints.PingEndpoint,
			decodeGRPCPingRequest,
			encodeGRPCPingResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "Ping", logger)))...,
		),
		getJobs: grpctransport.NewServer(
			endpoints.GetJobsEndpoint,
			decodeGRPCGetJobsRequest,
			encodeGRPCGetJobsResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "GetJobs", logger)))...,
		),
		newJob: grpctransport.NewServer(
			endpoints.NewJobEndpoint,
			decodeGRPCNewJobRequest,
			encodeGRPCNewJobResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "NewJob", logger)))...,
		),
	}
}

func (s *grpcServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingReply, error) {
	_, rep, err := s.ping.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.PingReply), nil
}

func (s *grpcServer) GetJobs(ctx context.Context, req *pb.GetJobsRequest) (*pb.GetJobsReply, error) {
	_, rep, err := s.getJobs.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.GetJobsReply), nil
}

func (s *grpcServer) NewJob(ctx context.Context, req *pb.NewJobRequest) (*pb.NewJobReply, error) {
	_, rep, err := s.newJob.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.NewJobReply), nil
}

// NewGRPCClient returns an WorkerService backed by a gRPC server at the other end
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
	// endpoint.Endpoint) that gets wrapped with various middlewares. If you
	// made your own client library, you'd do this work there, so your server
	// could rely on a consistent set of client behavior.
	var pingEndpoint kitendpoint.Endpoint
	{
		pingEndpoint = grpctransport.NewClient(
			conn,
			"pb.worker.Worker",
			"Ping",
			encodeGRPCPingRequest,
			decodeGRPCPingResponse,
			pb.PingReply{},
			append(options, grpctransport.ClientBefore(opentracing.ContextToGRPC(otTracer, logger)))...,
		).Endpoint()
		pingEndpoint = opentracing.TraceClient(otTracer, "Ping")(pingEndpoint)
		pingEndpoint = limiter(pingEndpoint)
		pingEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Ping",
			Timeout: 30 * time.Second,
		}))(pingEndpoint)
	}

	var getJobsEndpoint kitendpoint.Endpoint
	{
		getJobsEndpoint = grpctransport.NewClient(
			conn,
			"pb.worker.Worker",
			"GetJobs",
			encodeGRPCGetJobsRequest,
			decodeGRPCGetJobsResponse,
			pb.GetJobsReply{},
			append(options, grpctransport.ClientBefore(opentracing.ContextToGRPC(otTracer, logger)))...,
		).Endpoint()
		getJobsEndpoint = opentracing.TraceClient(otTracer, "GetJobs")(getJobsEndpoint)
		getJobsEndpoint = limiter(getJobsEndpoint)
		getJobsEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "GetJobs",
			Timeout: 30 * time.Second,
		}))(getJobsEndpoint)
	}

	var newJobEndpoint kitendpoint.Endpoint
	{
		newJobEndpoint = grpctransport.NewClient(
			conn,
			"pb.worker.Worker",
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
		PingEndpoint:    pingEndpoint,
		GetJobsEndpoint: getJobsEndpoint,
		NewJobEndpoint:  newJobEndpoint,
	}
}

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

// ********** Ping **********

// encodeGRPCPingRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain Ping request to a gRPC Ping request. Primarily useful in a client.
func encodeGRPCPingRequest(_ context.Context, request interface{}) (interface{}, error) {
	// req := request.(endpoint.PingRequest)
	return &pb.PingRequest{}, nil
}

// decodeGRPCPingRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC Ping request to a user-domain Ping request. Primarily useful in a server.
func decodeGRPCPingRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	// req := grpcReq.(*pb.PingRequest)
	return endpoint.PingRequest{}, nil
}

// encodeGRPCPingResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain Ping response to a gRPC Ping reply. Primarily useful in a server.
func encodeGRPCPingResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(endpoint.PingResponse)
	return &pb.PingReply{Jobs: int32(resp.JobsCount), Err: err2str(resp.Err)}, nil
}

// decodeGRPCPingResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC Ping reply to a user-domain Ping response. Primarily useful in a client.
func decodeGRPCPingResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.PingReply)
	return endpoint.PingResponse{JobsCount: int(reply.Jobs), Err: str2err(reply.Err)}, nil
}

// ********** GetJobs **********

// encodeGRPCGetJobsRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain GetJobs request to a gRPC GetJobs request. Primarily useful in a client.
func encodeGRPCGetJobsRequest(_ context.Context, request interface{}) (interface{}, error) {
	// req := request.(endpoint.GetJobsRequest)
	return &pb.GetJobsRequest{}, nil
}

// decodeGRPCGetJobsRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC GetJobs request to a user-domain GetJobs request. Primarily useful in a server.
func decodeGRPCGetJobsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	// req := grpcReq.(*pb.GetJobsRequest)
	return endpoint.GetJobsRequest{}, nil
}

// encodeGRPCGetJobsResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain GetJobs response to a gRPC GetJobs reply. Primarily useful in a server.
func encodeGRPCGetJobsResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(endpoint.GetJobsResponse)

	// conver []worker.Job to []*pb.Job
	pbJobs := make([]*pb.Job, 0)
	for _, j := range resp.Jobs {
		st, _ := timestamp.TimestampProto(j.StartTime)
		ft, _ := timestamp.TimestampProto(j.FinishTime)
		pbJob := &pb.Job{
			ID:         j.ID.String(),
			Per:        j.Per,
			Duration:   float32(j.Duration.Seconds()),
			StartTime:  st,
			FinishTime: ft,
		}
		pbJobs = append(pbJobs, pbJob)
	}
	return &pb.GetJobsReply{Jobs: pbJobs, Err: err2str(resp.Err)}, nil
}

// decodeGRPCGetJobsResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC GetJobs reply to a user-domain GetJobs response. Primarily useful in a client.
func decodeGRPCGetJobsResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.GetJobsReply)

	// conver []*pb.Job to []worker.Job
	jobs := make([]worker.Job, 0)
	for _, j := range reply.Jobs {
		id, _ := uuid.Parse(j.ID)
		dur, _ := time.ParseDuration(fmt.Sprintf("%f", j.Duration) + "s")
		st, _ := timestamp.Timestamp(j.StartTime)
		ft, _ := timestamp.Timestamp(j.FinishTime)
		job := worker.Job{
			ID:         worker.JobID{id},
			Per:        j.Per,
			Duration:   dur,
			StartTime:  st,
			FinishTime: ft,
		}
		jobs = append(jobs, job)
	}

	return endpoint.GetJobsResponse{Jobs: jobs, Err: str2err(reply.Err)}, nil
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
	return &pb.NewJobReply{Id: resp.ID, Err: err2str(resp.Err)}, nil
}

// decodeGRPCNewJobResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC NewJob reply to a user-domain NewJob response. Primarily useful in a client.
func decodeGRPCNewJobResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.NewJobReply)
	return endpoint.NewJobResponse{ID: reply.Id, Err: str2err(reply.Err)}, nil
}
