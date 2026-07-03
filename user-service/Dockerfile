FROM golang:1.26.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o main ./cmd/api

FROM alpine:3.23.2
RUN adduser -D -H -h /app appuser
WORKDIR /app
RUN chown appuser:appuser /app
COPY --from=builder --chown=appuser:appuser /app/main .
USER appuser
EXPOSE 9090
CMD ["./main"]
