FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/search-service cmd/search-service/main.go

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache bash postgresql-client

COPY --from=build /app/search-service /app/search-service
COPY --from=build /app/config /app/config
COPY --from=build /app/.env /app/.env

EXPOSE 9091

CMD ["/app/search-service"]