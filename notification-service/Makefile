fmt:
	go fmt ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

test:
	go run tests/send_test_message.go

up:
	docker compose up -d

up-build:
	docker compose up -d --build

down:
	docker compose down