FROM golang:1.25.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY gen/calculatorpb/ ./gen/calculatorpb/
COPY services/calculator/ ./services/calculator/
RUN CGO_ENABLED=0 GOOS=linux go build -o /calculator ./services/calculator/cmd/main.go

FROM alpine:latest AS probe
RUN wget -qO /grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.24/grpc_health_probe-linux-amd64 && \
    chmod +x /grpc_health_probe

FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
COPY --from=builder /calculator /calculator
COPY --from=probe /grpc_health_probe /usr/local/bin/grpc_health_probe
EXPOSE 50051
CMD ["/calculator"]