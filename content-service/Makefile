default: run

run:
	go run cmd/content-service/main.go

fmt:
	go fmt ./...

lint:
	golangci-lint run

migrate:
	go run cmd/migrate/main.go

test-unit tu:
	go test ./internal/services/...

test-functional tf:
	make ctc
	go test ./tests
	make dtc

create-test-content ctc:
	docker exec -i content-service-db psql -U postgres -d content-service-db < tests/suite/create_test_content.sql

delete-test-content dtc:
	docker exec -i content-service-db psql -U postgres -d content-service-db < tests/suite/delete_test_content.sql

clear-db:
	docker exec -i content-service-db psql -U postgres -d content-service-db < tests/suite/clear_db.sql
