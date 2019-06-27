FROM golang:alpine

EXPOSE 8080
EXPOSE 8081

ADD ./ /go/src/apiserver
WORKDIR /go/src/apiserver

RUN go install -v ./cmd/apiserver

ENTRYPOINT ["apiserver"]