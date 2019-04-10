# Transaction App
**Transaction App** is a simple microservice application aims to show how modern distributed application can be builded and deployed.
It is written on Go programming language. 
It demonstrate how to build microservice that exposes its operation over over all three major transports: GRPC, REST API, Messaging System (NATS). It also shows how to do logging and distributed tracing to external systems and expose application metrics for external monitoring. 

**Transaction App** can be deployed on a single docker host or on a Kubernetes cluster. See [Deploy](https://github.com/Maxfer4Maxfer/transaction_app/blob/master/README.md#deploy) section.

**Transaction App** is a simple application. From user interface you can only run one or batch of new jobs. Application desided on which worker the new job should be runned. The more jobs started the slower they are performed. Application shows you progress of each jobs and displays information of already completed jobs. 

![Transaction App Overview Diagram](https://github.com/Maxfer4Maxfer/transaction_app/blob/master/docs/pics/Diagrams-Overview.jpg)

## Technology stack
Programming Languages:
* [GO](https://golang.org) for the apiserver, repository and worker components.
* [React](https://reactjs.org) for  building user interface.

Programming toolkit:
* [go-kit](https://gokit.io) for building a microservice architecture.
* [gorm](http://gorm.io) ORM library for making interaction with a database.
* [GRPC](https://grpc.io) a high-performance, open-source universal RPC framework. Interaction between all microservice. 
* [protobuf](https://github.com/golang/protobuf) Google's data interchange format.
* [REST API](https://en.wikipedia.org/wiki/Representational_state_transfer) Interaction between ui and apiserver components. Request: http. Response: json.
* [OpenTracing](https://opentracing.io) instrumentation for connecting to a distributed tracing system.

Third party software components:
* [MySQL](https://www.mysql.com) the world know relational database.
* [NATS](https://nats.io) a high performance messaging system.
* [Prometheus](https://prometheus.io) monitoring system for gathering application, docker and operation systems metrics .
* [Grafana](https://grafana.com) for visualisation metrics from Prometheus.
* [ELK](https://www.elastic.co/elk-stack) for gathering docker and operation systems logs. ELK stack: Filebeat -> Logstash -> ElasticSearch -> Kibana.
* [Jaeger](http://jaegertracing.io) for gather, store and analyse trace from application.

## Deploy
There are three options for deploying **Transaction App**
* [Simple docker deploy](https://github.com/Maxfer4Maxfer/transaction_app/blob/master/docs/pics/Diagrams-Overview.jpg)
* [Docker compose deploy](https://github.com/Maxfer4Maxfer/transaction_app/blob/master/docs/pics/Diagrams-Overview.jpg)
* [Kubernetes deploy](https://github.com/Maxfer4Maxfer/transaction_app/blob/master/docs/pics/Diagrams-Overview.jpg)

## Entrypoints
* http:__*<<ip_address>>*__:80 Transaction App's user interface
* http:__*<<ip_address>>*__:9090 Prometheus
* http:__*<<ip_address>>*__:3000 Grafana
* http:__*<<ip_address>>*__:5601 Kibana
* http:__*<<ip_address>>*__:3000 Jaeger

## Components specification
ui

![ui](https://github.com/Maxfer4Maxfer/transaction_app/blob/master/docs/pics/Diagrams-ui.jpg)

apiserver

![appserver](https://github.com/Maxfer4Maxfer/transaction_app/blob/master/docs/pics/Diagrams-apiserver.jpg)

repository

![repository](https://github.com/Maxfer4Maxfer/transaction_app/blob/master/docs/pics/Diagrams-repository.jpg)

worker

![worker](https://github.com/Maxfer4Maxfer/transaction_app/blob/master/docs/pics/Diagrams-worker.jpg)


## Donations
 If you want to support this project, please consider donating:
 * PayPal: https://paypal.me/MaxFe
