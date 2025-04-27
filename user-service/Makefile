fmt:
	go fmt ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

migrate:
	go run cmd/migrate/main.go

test-unit tu:
	go test ./internal/services/...

test-functional tf:
	make ctu
	go test ./tests
	make dtu

create-test-users ctu:
	docker exec -i user-service-db psql -U postgres -d user-service-db < tests/suite/create_test_users.sql

delete-test-users dtu:
	docker exec -i user-service-db psql -U postgres -d user-service-db < tests/suite/delete_test_users.sql

clear-db:
	docker exec -i user-service-db psql -U postgres -d user-service-db < tests/suite/clear_db.sql

up:
	docker compose up -d

up-build:
	docker compose up -d --build

down:
	docker compose down