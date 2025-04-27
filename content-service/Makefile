fmt:
	go fmt ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

migrate:
	go run cmd/migrate/main.go

up:
	docker compose up -d

up-build:
	docker compose up -d --build

down:
	docker compose down