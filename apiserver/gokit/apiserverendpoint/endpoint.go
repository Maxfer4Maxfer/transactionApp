package apiserverendpoint

import (
	"context"
	"time"

	stdopentracing "github.com/opentracing/opentracing-go"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"

	"apiserver/gokit/apiserverservice"
	"repository/repo"
)

// EndpointSet collects all of the endpoints that compose a apiserver service. It's meant to
// be used as a helper struct, to collect all of the endpoints into a single
// parameter.
type EndpointSet struct {
	GetAllNodesEndpoint endpoint.Endpoint
	NewJobEndpoint      endpoint.Endpoint
}

// New returns a Set that wraps the provided server, and wires in all of the
// expected endpoint middlewares via the various parameters.
func New(svc apiserverservice.Service, logger log.Logger, duration metrics.Histogram, otTracer stdopentracing.Tracer) EndpointSet {

	var getAllNodesEndpoint endpoint.Endpoint
	{
		getAllNodesEndpoint = MakeGetAllNodesEndpoint(svc)
		getAllNodesEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Millisecond), 1))(getAllNodesEndpoint)
		getAllNodesEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(getAllNodesEndpoint)
		getAllNodesEndpoint = opentracing.TraceServer(otTracer, "GetAllNodes")(getAllNodesEndpoint)
		getAllNodesEndpoint = LoggingMiddleware(log.With(logger, "method", "GetAllNodes"))(getAllNodesEndpoint)
		getAllNodesEndpoint = InstrumentingMiddleware(duration.With("method", "GetAllNodes"))(getAllNodesEndpoint)
	}

	var newJobEndpoint endpoint.Endpoint
	{
		newJobEndpoint = MakeNewJobEndpoint(svc)
		newJobEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Millisecond), 1))(newJobEndpoint)
		newJobEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(newJobEndpoint)
		newJobEndpoint = opentracing.TraceServer(otTracer, "NewJob")(newJobEndpoint)
		newJobEndpoint = LoggingMiddleware(log.With(logger, "method", "NewJob"))(newJobEndpoint)
		newJobEndpoint = InstrumentingMiddleware(duration.With("method", "NewJob"))(newJobEndpoint)
	}

	return EndpointSet{
		GetAllNodesEndpoint: getAllNodesEndpoint,
		NewJobEndpoint:      newJobEndpoint,
	}
}

// ========= GetAllNodes ===========

// GetAllNodes implements the service interface, so EndpointSet may be used as a service.
// This is primarily useful in the context of a client library.
func (s EndpointSet) GetAllNodes(ctx context.Context) ([]repo.Node, error) {
	resp, err := s.GetAllNodesEndpoint(ctx, GetAllNodesRequest{})
	if err != nil {
		return nil, err
	}
	response := resp.(GetAllNodesResponse)
	return response.Nodes, response.Err
}

// GetAllNodesRequest collects the request parameters for the GetAllNodes method.
type GetAllNodesRequest struct {
}

// GetAllNodesResponse collects the response values for the GetAllNodes method.
type GetAllNodesResponse struct {
	Nodes []repo.Node `json:"nodes"`
	Err   error       `json:"-"` // should be intercepted by Failed/errorEncoder
}

// MakeGetAllNodesEndpoint constructs a Sum endpoint wrapping the service.
func MakeGetAllNodesEndpoint(s apiserverservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		// req := request.(GetAllNodesRequest)
		nodes, err := s.GetAllNodes(ctx)
		return GetAllNodesResponse{Nodes: nodes, Err: err}, nil
	}
}

// ========= NewJob ===========

// NewJob implements the service interface, so EndpointSet may be used as a service.
// This is primarily useful in the context of a client library.
func (s EndpointSet) NewJob(ctx context.Context) (string, error) {
	resp, err := s.NewJobEndpoint(ctx, NewJobRequest{})
	if err != nil {
		return "-1", err
	}
	response := resp.(NewJobResponse)
	return response.ID, response.Err
}

// NewJobRequest collects the request parameters for the NewJob method.
type NewJobRequest struct {
}

// NewJobResponse collects the response values for the NewJob method.
type NewJobResponse struct {
	ID  string `json:"id"`
	Err error  `json:"-"` // should be intercepted by Failed/errorEncoder
}

// MakeNewJobEndpoint constructs a Sum endpoint wrapping the service.
func MakeNewJobEndpoint(s apiserverservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		// req := request.(NewJobRequest)
		id, err := s.NewJob(ctx)
		return NewJobResponse{ID: id, Err: err}, nil
	}
}
