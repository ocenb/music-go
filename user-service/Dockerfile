FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/user-service cmd/user-service/main.go
RUN go build -o /app/migrate cmd/migrate/main.go

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache bash postgresql-client

COPY --from=build /app/user-service /app/user-service
COPY --from=build /app/migrate /app/migrate
COPY --from=build /app/config /app/config
COPY --from=build /app/.env /app/.env
COPY --from=build /app/migrations /app/migrations

COPY --from=build /app/start.sh /app/start.sh
RUN chmod +x /app/start.sh

EXPOSE 9090

CMD ["/app/start.sh"]