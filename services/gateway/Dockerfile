FROM golang:1.25.1-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY gen/calculatorpb/ ./gen/calculatorpb/
COPY services/gateway/ ./services/gateway/

RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway ./services/gateway/cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /gateway /gateway
EXPOSE 8080
CMD ["/gateway"]