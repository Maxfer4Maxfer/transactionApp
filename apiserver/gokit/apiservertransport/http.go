package apiservertransport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	stdopentracing "github.com/opentracing/opentracing-go"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	httptransport "github.com/go-kit/kit/transport/http"

	"apiserver/gokit/apiserverendpoint"
	"apiserver/gokit/apiserverservice"
)

// NewHTTPHandler returns an HTTP handler that makes a set of endpoints
// available on predefined paths.
func NewHTTPHandler(endpoints apiserverendpoint.EndpointSet, otTracer stdopentracing.Tracer, logger log.Logger) http.Handler {

	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorLogger(logger),
	}

	m := http.NewServeMux()
	m.Handle("/getallnodes", httptransport.NewServer(
		endpoints.GetAllNodesEndpoint,
		decodeHTTPGetAllNodesRequest,
		encodeHTTPGenericResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetAllNodes", logger)))...,
	))
	m.Handle("/newjob", httptransport.NewServer(
		endpoints.NewJobEndpoint,
		decodeHTTPNewJobRequest,
		encodeHTTPGenericResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "NewJob", logger)))...,
	))
	return accessControl(m)
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(err2code(err))
	json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}

func err2code(err error) int {
	switch err {
	case apiserverservice.ErrAPIServerUnevailable:
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func errorDecoder(r *http.Response) error {
	var w errorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Error)
}

type errorWrapper struct {
	Error string `json:"error"`
}

// encodeHTTPGenericRequest is a transport/http.EncodeRequestFunc that
// JSON-encodes any request to the request body. Primarily useful in a client.
func encodeHTTPGenericRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

// encodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func encodeHTTPGenericResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
		errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// ======= GetAllNodes ======

// decodeHTTPGetAllNodesRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded GetAllNodes request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPGetAllNodesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req apiserverendpoint.GetAllNodesRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// decodeHTTPGetAllNodesResponse is a transport/http.DecodeResponseFunc that decodes
// a JSON-encoded GetAllNodes response from the HTTP response body. If the response
// has a non-200 status code, we will interpret that as an error and attempt to
// decode the specific error message from the response body. Primarily useful in
// a client.
func decodeHTTPGetAllNodesResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp apiserverendpoint.GetAllNodesResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

// ======= NewJob ======

// decodeHTTPNewJobRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded NewJob request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPNewJobRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req apiserverendpoint.NewJobRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// decodeHTTPNewJobResponse is a transport/http.DecodeResponseFunc that decodes
// a JSON-encoded NewJob response from the HTTP response body. If the response
// has a non-200 status code, we will interpret that as an error and attempt to
// decode the specific error message from the response body. Primarily useful in
// a client.
func decodeHTTPNewJobResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp apiserverendpoint.NewJobResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}
