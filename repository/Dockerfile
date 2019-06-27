FROM golang:alpine

EXPOSE 8080
EXPOSE 8082

ADD ./ /go/src/repository
WORKDIR /go/src/repository

RUN apk add git
RUN go mod download
RUN go install -v ./cmd/repository

ENTRYPOINT ["repository"]
CMD ["--nats-addr", "nats:4222"]
