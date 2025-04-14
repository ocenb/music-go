FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/content-service cmd/content-service/main.go
RUN go build -o /app/migrate cmd/migrate/main.go

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache bash postgresql-client ffmpeg

RUN mkdir -p /app/temp && chmod 777 /app/temp

COPY --from=build /app/content-service /app/content-service
COPY --from=build /app/migrate /app/migrate
COPY --from=build /app/config /app/config
COPY --from=build /app/.env /app/.env
COPY --from=build /app/migrations /app/migrations

COPY --from=build /app/start.sh /app/start.sh
RUN chmod +x /app/start.sh

EXPOSE 3000

CMD ["/app/start.sh"]