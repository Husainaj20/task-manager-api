# Multi-stage build: static binary with musl
FROM golang:1.22-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org,direct
RUN go mod download
COPY . ./
# build a static, stripped binary
RUN apk add --no-cache build-base && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /app/task-manager-api ./cmd/server

# Use distroless-ish minimal base (scratch + ca certs)
FROM gcr.io/distroless/static
COPY --from=build /app/task-manager-api /usr/local/bin/task-manager-api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/task-manager-api"]
# multi-stage Go build
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org,direct
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /app/cmd/server ./cmd/server

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /app/cmd/server /usr/local/bin/task-manager-api
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/task-manager-api"]
