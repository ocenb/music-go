FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/notification-service cmd/notification-service/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/notification-service /app/notification-service
COPY --from=build /app/config /app/config
COPY --from=build /app/.env /app/.env

EXPOSE 9089

CMD ["/app/notification-service"]