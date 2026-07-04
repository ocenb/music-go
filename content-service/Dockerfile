FROM golang:1.26.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o main ./cmd/content-service

FROM alpine:3.23.2
RUN apk add --no-cache ffmpeg
RUN adduser -D -H -h /app appuser && mkdir -p /app/temp && chown appuser:appuser /app/temp
WORKDIR /app
RUN chown appuser:appuser /app
COPY --from=builder --chown=appuser:appuser /app/main .
COPY --from=builder --chown=appuser:appuser /app/config ./config
USER appuser
EXPOSE 3000
CMD ["./main"]
