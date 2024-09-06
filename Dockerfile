ARG GO_VERSION=1.21.11

FROM golang:${GO_VERSION}-alpine AS builder

RUN go env -w GOPROXY=direct
RUN apk add --no-cache git
RUN apk --no-cache add ca-certificates && update-ca-certificates

WORKDIR /src

COPY ./go.mod ./go.sum ./

RUN go mod download 

COPY models models
COPY events events
COPY repository repository
COPY database database
COPY search search
COPY feed-service feed-service
COPY query-service query-service
COPY pusher-service pusher-service

RUN go install ./...

FROM alpine:3.17
WORKDIR /usr/bin

COPY --from=builder /go/bin .
