FROM golang:alpine

EXPOSE 8080
EXPOSE 8082

ADD ./ /go/src/worker
WORKDIR /go/src/worker

RUN go install -v ./cmd/worker

ENTRYPOINT ["worker"]
CMD ["--nats-addr", "nats:4222"]
