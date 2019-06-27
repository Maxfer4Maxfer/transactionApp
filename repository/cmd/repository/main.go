package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/nats-io/go-nats"
	"github.com/oklog/oklog/pkg/group"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	jaegercfg "github.com/uber/jaeger-client-go/config"

	"google.golang.org/grpc"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	grpctransport "github.com/go-kit/kit/transport/grpc"

	"repository/pkg/endpoint"
	"repository/pkg/service"
	"repository/pkg/transport"
	store "repository/pkg/storage/gorm"
	repopb "repository/pb"
)

func main() {
	fs := flag.NewFlagSet("reposvc", flag.ExitOnError)
	var (
		debugAddr = fs.String("debug-addr", ":8080", "Debug and metrics listen address")
		grpcAddr  = fs.String("grpc-addr", ":8082", "gRPC listen address")
		natsAddr  = fs.String("nats-addr", nats.DefaultURL, "NATS server address")
		dsn       = fs.String("dsn", "root:root@tcp(localhost:3306)/cache?charset=utf8&parseTime=True&loc=Local", "Database Source Name")
		jaegerURL = fs.String("jaeger-addr", "jaeger:5775", "Jaeger server address")
	)

	fs.Usage = usageFor(fs, os.Args[0]+" [flags]")
	fs.Parse(os.Args[1:])

	// Create a single logger, which we'll use and give to other components.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	// We'll pass the tracer to all the
	// components that use it, as a dependency.
	var tracer stdopentracing.Tracer
	{
		cfg := jaegercfg.Configuration{
			Sampler: &jaegercfg.SamplerConfig{
				Type:  "const",
				Param: 1,
			},
			Reporter: &jaegercfg.ReporterConfig{
				LogSpans:            true,
				BufferFlushInterval: 1 * time.Second,
				LocalAgentHostPort:  *jaegerURL,
			},
		}

		closer, err := cfg.InitGlobalTracer("Repo")
		if err != nil {
			panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
		}
		logger.Log("tracer", "Jaeger", "type", "OpenTracing", "URL", *jaegerURL)
		tracer = stdopentracing.GlobalTracer()
		defer closer.Close()
	}

	// Create the (sparse) metrics we'll use in the service. They, too, are
	// dependencies that we pass to components that use them.
	var registerNodes, getAllNodes, newJobs metrics.Counter
	{
		// Business-level metrics.
		registerNodes = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "transactionApp",
			Subsystem: "repository",
			Name:      "nodes_registered",
			Help:      "Total count of nodes registored via the RegisterNodes method.",
		}, []string{})
		getAllNodes = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "transactionApp",
			Subsystem: "repository",
			Name:      "getallnodes_called",
			Help:      "Total count the GetAllNodes method called.",
		}, []string{})
		newJobs = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "transactionApp",
			Subsystem: "repository",
			Name:      "new_jobs",
			Help:      "Total count new jobs started via the NewJob method.",
		}, []string{})
	}
	var duration metrics.Histogram
	{
		// Endpoint-level metrics.
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "transactionApp",
			Subsystem: "repository",
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds.",
		}, []string{"method", "success"})
	}
	http.DefaultServeMux.Handle("/metrics", promhttp.Handler())

	// Build the layers of the service "onion" from the inside out. First, the
	// business logic service; then, the set of endpoints that wrap the service;
	// and finally, a series of concrete transport adapters. The adapters, like
	// the HTTP handler or the gRPC server, are the bridge between Go kit and
	// the interfaces that the transports expect. Note that we're not binding
	// them to ports or anything yet; we'll do that next.
	var (
		storage, sCloser = store.New(*dsn, logger)
		service          = service.New(storage, logger, registerNodes, getAllNodes, newJobs)
		endpoints        = endpoint.New(service, logger, duration, tracer)
		natsSubscribers  = transport.NewNATSSubscribers(endpoints, tracer, logger)
		grpcServer       = transport.NewGRPCServer(endpoints, tracer, logger)
	)
	defer sCloser()

	natsCloser := transport.NewNATSHandler(*natsAddr, natsSubscribers, logger)
	defer natsCloser()

	// Now we're to the part of the func main where we want to start actually
	// running things, like servers bound to listeners to receive connections.
	//
	// The method is the same for each component: add a new actor to the group
	// struct, which is a combination of 2 anonymous functions: the first
	// function actually runs the component, and the second function should
	// interrupt the first function and cause it to return. It's in these
	// functions that we actually bind the Go kit server/handler structs to the
	// concrete transports and run them.
	//
	// Putting each component into its own block is mostly for aesthetics: it
	// clearly demarcates the scope in which each listener/socket may be used.
	var g group.Group
	{
		// The debug listener mounts the http.DefaultServeMux, and serves up
		// stuff like the Prometheus metrics route, the Go debug and profiling
		// routes, and so on.
		debugListener, err := net.Listen("tcp", *debugAddr)
		if err != nil {
			logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "debug/HTTP", "addr", *debugAddr)
			return http.Serve(debugListener, http.DefaultServeMux)
		}, func(error) {
			debugListener.Close()
		})
	}
	{
		// The gRPC listener mounts the Go kit gRPC server we created.
		grpcListener, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			logger.Log("transport", "gRPC", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "gRPC", "addr", *grpcAddr)
			// we add the Go Kit gRPC Interceptor to our gRPC service as it is used by
			// the here demonstrated zipkin tracing middleware.
			baseServer := grpc.NewServer(grpc.UnaryInterceptor(grpctransport.Interceptor))
			repopb.RegisterRepoServer(baseServer, grpcServer)
			return baseServer.Serve(grpcListener)
		}, func(error) {
			grpcListener.Close()
		})
	}
	{
		// This function just sits and waits for ctrl-C.
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}
	logger.Log("exit", g.Run())
}

func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}
