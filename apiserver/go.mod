module apiserver

go 1.12

require (
	github.com/go-kit/kit v0.8.0
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/oklog/oklog v0.3.2
	github.com/opentracing/opentracing-go v1.1.0
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/sony/gobreaker v0.0.0-20190329013020-a9b2a3fc7395
	github.com/uber/jaeger-client-go v2.16.0+incompatible
	github.com/uber/jaeger-lib v2.0.0+incompatible // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	google.golang.org/grpc v1.19.1
	repository v0.0.0
)

replace repository => ../repository

replace worker => ../worker
