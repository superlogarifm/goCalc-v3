FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /orchestrator ./cmd/orchestrator

FROM alpine:latest

WORKDIR /app
COPY --from=builder /orchestrator .

EXPOSE 8080
CMD ["./orchestrator"] 