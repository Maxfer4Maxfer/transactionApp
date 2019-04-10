module repository

require (
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/go-kit/kit v0.8.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/protobuf v1.3.0
	github.com/google/uuid v1.1.1
	github.com/gopherjs/gopherjs v0.0.0-20181103185306-d547d1d9531e // indirect
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jtolds/gls v4.2.1+incompatible // indirect
	github.com/nats-io/gnatsd v1.4.1 // indirect
	github.com/nats-io/go-nats v1.7.2
	github.com/oklog/oklog v0.3.2
	github.com/opentracing/opentracing-go v1.0.2
	github.com/prometheus/client_golang v0.9.2
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90 // indirect
	github.com/prometheus/common v0.2.0 // indirect
	github.com/prometheus/procfs v0.0.0-20190209105433-f8d8b3f739bd // indirect
	github.com/smartystreets/assertions v0.0.0-20190215210624-980c5ac6f3ac // indirect
	github.com/smartystreets/goconvey v0.0.0-20181108003508-044398e4856c // indirect
	github.com/sony/gobreaker v0.0.0-20181109014844-d928aaea92e1
	github.com/uber/jaeger-client-go v2.15.0+incompatible
	golang.org/x/crypto v0.0.0-20190211182817-74369b46fc67 // indirect
	golang.org/x/net v0.0.0-20190213061140-3a22650c66bd // indirect
	golang.org/x/sys v0.0.0-20190215142949-d0b11bdaac8a // indirect
	golang.org/x/text v0.3.1-0.20180807135948-17ff2d5776d2 // indirect
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c
	google.golang.org/genproto v0.0.0-20190215211957-bd968387e4aa // indirect
	google.golang.org/grpc v1.19.0
	worker v0.0.0
)

replace worker => ../worker
