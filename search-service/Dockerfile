FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -a -installsuffix cgo -o /app/search-service cmd/search-service/main.go

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache upgrade

COPY --from=build /app/search-service /app/search-service

EXPOSE 9091

CMD ["/app/search-service"]