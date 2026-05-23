# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -o /go/bin/goose github.com/pressly/goose/v3/cmd/goose

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /server /app/server
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations ./migrations

ENV SERVER_PORT=8080
EXPOSE 8080
CMD ["/app/server"]
