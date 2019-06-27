package endpoint

import (
	"context"
	"time"

	stdopentracing "github.com/opentracing/opentracing-go"

	"github.com/go-kit/kit/circuitbreaker"
	kitendpoint "github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"

	"worker/pkg/service"
	worker "worker/pkg/model"
)

// EndpointSet collects all of the endpoints that compose a repo service. It's meant to
// be used as a helper struct, to collect all of the endpoints into a single
// parameter.
type EndpointSet struct {
	PingEndpoint    kitendpoint.Endpoint
	NewJobEndpoint  kitendpoint.Endpoint
	GetJobsEndpoint kitendpoint.Endpoint
}

// New returns a Set that wraps the provided server, and wires in all of the
// expected endpoint middlewares via the various parameters.
func New(svc service.Service, logger log.Logger, duration metrics.Histogram, otTracer stdopentracing.Tracer) EndpointSet {
	var pingEndpoint kitendpoint.Endpoint
	{
		pingEndpoint = MakePingEndpoint(svc)
		pingEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Millisecond), 1))(pingEndpoint)
		pingEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(pingEndpoint)
		pingEndpoint = opentracing.TraceServer(otTracer, "Ping")(pingEndpoint)
		pingEndpoint = LoggingMiddleware(log.With(logger, "method", "Ping"))(pingEndpoint)
		pingEndpoint = InstrumentingMiddleware(duration.With("method", "Ping"))(pingEndpoint)
	}

	var newJobEndpoint kitendpoint.Endpoint
	{
		newJobEndpoint = MakeNewJobEndpoint(svc)
		newJobEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Millisecond), 1))(newJobEndpoint)
		newJobEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(newJobEndpoint)
		newJobEndpoint = opentracing.TraceServer(otTracer, "NewJob")(newJobEndpoint)
		newJobEndpoint = LoggingMiddleware(log.With(logger, "method", "NewJob"))(newJobEndpoint)
		newJobEndpoint = InstrumentingMiddleware(duration.With("method", "NewJob"))(newJobEndpoint)
	}

	var getJobsEndpoint kitendpoint.Endpoint
	{
		getJobsEndpoint = MakeGetJobsEndpoint(svc)
		getJobsEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Millisecond), 1))(getJobsEndpoint)
		getJobsEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(getJobsEndpoint)
		getJobsEndpoint = opentracing.TraceServer(otTracer, "GetJobs")(getJobsEndpoint)
		getJobsEndpoint = LoggingMiddleware(log.With(logger, "method", "GetJobs"))(getJobsEndpoint)
		getJobsEndpoint = InstrumentingMiddleware(duration.With("method", "GetJobs"))(getJobsEndpoint)
	}

	return EndpointSet{
		PingEndpoint:    pingEndpoint,
		NewJobEndpoint:  newJobEndpoint,
		GetJobsEndpoint: getJobsEndpoint,
	}
}

// ================ Ping =============

// Ping implements the service interface, so EndpointSet may be used as a service.
// This is primarily useful in the context of a client library.
func (s EndpointSet) Ping(ctx context.Context) (int, error) {
	resp, err := s.PingEndpoint(ctx, PingRequest{})
	if err != nil {
		return 0, err
	}
	response := resp.(PingResponse)
	return response.JobsCount, response.Err
}

// PingRequest collects the request parameters for the Ping method.
type PingRequest struct {
}

// PingResponse collects the response values for the Ping method.
type PingResponse struct {
	JobsCount int   `json:"jobsCount"`
	Err       error `json:"-"` // should be intercepted by Failed/errorEncoder
}

// MakePingEndpoint constructs a Sum endpoint wrapping the service.
func MakePingEndpoint(s service.Service) kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		// req := request.(PingRequest)
		jobscount, err := s.Ping(ctx)
		return PingResponse{JobsCount: jobscount, Err: err}, nil
	}
}

// ================ NewJob =============

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
func MakeNewJobEndpoint(s service.Service) kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		// req := request.(NewJobRequest)
		id, err := s.NewJob(ctx)
		return NewJobResponse{ID: id, Err: err}, nil
	}
}

// ================ GetJobs =============

// GetJobs implements the service interface, so EndpointSet may be used as a service.
// This is primarily useful in the context of a client library.
func (s EndpointSet) GetJobs(ctx context.Context) ([]worker.Job, error) {
	resp, err := s.GetJobsEndpoint(ctx, GetJobsRequest{})
	if err != nil {
		return nil, err
	}
	response := resp.(GetJobsResponse)
	return response.Jobs, response.Err
}

// GetJobsRequest collects the request parameters for the GetJobs method.
type GetJobsRequest struct {
}

// GetJobsResponse collects the response values for the GetJobs method.
type GetJobsResponse struct {
	Jobs []worker.Job `json:"jobs"`
	Err  error        `json:"-"` // should be intercepted by Failed/errorEncoder
}

// MakeGetJobsEndpoint constructs a Sum endpoint wrapping the service.
func MakeGetJobsEndpoint(s service.Service) kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		// req := request.(GetJobsRequest)
		jobs, err := s.GetJobs(ctx)
		return GetJobsResponse{Jobs: jobs, Err: err}, nil
	}
}
