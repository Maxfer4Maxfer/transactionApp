package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/go-nats"
	stdopentracing "github.com/opentracing/opentracing-go"

	"github.com/go-kit/kit/log"
	natstransport "github.com/go-kit/kit/transport/nats"

	"repository/pkg/endpoint"
)

type NATSSubscribers map[string]*natstransport.Subscriber

// NewNATSSubscribers returns an NATS subscribers that makes a set of endpoints
// available on predefined paths.
func NewNATSSubscribers(endpoint endpoint.EndpointSet, otTracer stdopentracing.Tracer, logger log.Logger) NATSSubscribers {

	var subscribers NATSSubscribers = make(map[string]*natstransport.Subscriber)

	RegisterNodeHandler := natstransport.NewSubscriber(
		endpoint.RegisterNodeEndpoint,
		decodeNATSRegisterNodeRequest,
		EncodeJSONResponse,
		natstransport.SubscriberBefore(func(ctx context.Context, msg *nats.Msg) context.Context { return ctx }),
		natstransport.SubscriberAfter(func(ctx context.Context, nc *nats.Conn) context.Context { return ctx }),
	)

	subscribers["RegisterNode"] = RegisterNodeHandler

	return subscribers

}

// decodeHTTPRegisterNodeRequest is a transport/nats.DecodeRequestFunc that decodes a
// JSON-encoded sum request from the NATS request body. Primarily useful in a
// server.
func decodeNATSRegisterNodeRequest(_ context.Context, m *nats.Msg) (interface{}, error) {
	var req endpoint.RegisterNodeRequest
	r := bytes.NewReader(m.Data)
	err := json.NewDecoder(r).Decode(&req)
	return req, err
}

type errorWrapper struct {
	Error string `json:"err"`
}

// EncodeJSONResponse is a EncodeResponseFunc that serializes the response as a
// JSON object to the subscriber reply. Many JSON-over services can use it as
// a sensible default.
func EncodeJSONResponse(_ context.Context, reply string, nc *nats.Conn, response interface{}) error {
	var err error
	var b []byte
	resp := response.(endpoint.RegisterNodeResponse)

	if resp.Err != nil {
		b, err = json.Marshal(errorWrapper{Error: resp.Err.Error()})
	} else {
		b, err = json.Marshal(resp)
	}
	if err != nil {
		return err
	}

	return nc.Publish(reply, b)
}

type NATSHandler struct {
	nc            *nats.Conn
	natsAddr      string
	subscribers   NATSSubscribers
	subscriptions []*nats.Subscription
	logger        log.Logger
}

// New create in memory repository for storing nodes
func NewNATSHandler(natsAddr string, subs NATSSubscribers, logger log.Logger) func() {

	nh := &NATSHandler{
		nc:            nil,
		natsAddr:      natsAddr,
		subscribers:   subs,
		subscriptions: make([]*nats.Subscription, 0),
		logger:        logger,
	}

	closeCh := make(chan struct{}, 1)
	go nh.connectToNATS(closeCh)
	closeFn := func() {
		closeCh <- struct{}{}
	}

	return closeFn
}

func (nh *NATSHandler) connectToNATS(closeCh chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			if nh.nc == nil {
				nh.logger.Log("transport", "NATS", "message", "trying connect to the NATS...")
				nc, err := nats.Connect(nh.natsAddr)
				if err != nil {
					nh.logger.Log("transport", "NATS", "message", "got an error", "err", err)
				} else {
					nh.logger.Log("transport", "NATS", "message", "connection to NATS is established")
					nh.nc = nc
					for key, s := range nh.subscribers {
						sub, err := nc.QueueSubscribe(key, "Repository", s.ServeMsg(nh.nc))
						if err != nil {
							nh.logger.Log("transport", "NATS", "message", err)
						}
						nh.subscriptions = append(nh.subscriptions, sub)
					}
				}

			} else {
				if !nh.nc.IsConnected() {
					nh.logger.Log("transport", "NATS", "message", "lost connection to the NATS")
					nh.nc = nil
				}
			}
		}
		for _, s := range nh.subscriptions {
			s.Unsubscribe()
		}
		nh.nc.Close()
	}()
	<-closeCh
	ticker.Stop()
}
