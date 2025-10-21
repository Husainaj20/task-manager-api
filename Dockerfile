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

# Distroless runtime stage (used by CI when building with --target runtime-distroless)
FROM gcr.io/distroless/static AS runtime-distroless
COPY --from=build /app/task-manager-api /usr/local/bin/task-manager-api
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/task-manager-api"]

# Alpine runtime stage (default local runtime, includes curl for healthchecks and debugging)
FROM alpine:3.18 AS runtime-alpine
RUN apk add --no-cache ca-certificates curl
COPY --from=build /app/task-manager-api /usr/local/bin/task-manager-api
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/task-manager-api"]
