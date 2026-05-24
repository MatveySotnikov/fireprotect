FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY gen/pdfpb/ ./gen/pdfpb/
COPY services/pdfservice/ ./services/pdfservice/

RUN CGO_ENABLED=0 GOOS=linux go build -o /pdfservice ./services/pdfservice/cmd/main.go

# Скачиваем grpc_health_probe
FROM alpine:latest AS probe
RUN wget -qO /grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.24/grpc_health_probe-linux-amd64 && \
    chmod +x /grpc_health_probe

FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
COPY --from=builder /pdfservice /pdfservice
COPY --from=probe /grpc_health_probe /usr/local/bin/grpc_health_probe
EXPOSE 50052
CMD ["/pdfservice"]