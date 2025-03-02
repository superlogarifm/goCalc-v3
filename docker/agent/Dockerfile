FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /agent ./cmd/agent

FROM alpine:latest

WORKDIR /app
COPY --from=builder /agent .

CMD ["./agent"] 