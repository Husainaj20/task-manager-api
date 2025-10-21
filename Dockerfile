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
