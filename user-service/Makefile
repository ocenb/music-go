.PHONY: build fix lint migrate gen-mocks test-unit tu test-functional tf \
        create-test-users ctu delete-test-users dtu clear-db \
        run up up-dev up-build up-dev-build down prepare infra infra-dev logs test-cover

ifneq (,$(wildcard .env))
    include .env
    export
endif

DB_URL := postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

build:
	go build -o .build/user-service ./cmd/user-service/main.go

fix:
	golangci-lint run --fix

lint:
	golangci-lint run

migrate:
	goose -dir ./migrations postgres "$(DB_URL)" up

gen-mocks:
	mockery

test-unit tu:
	go test ./internal/...

test-cover:
	go test -cover ./internal/...

test-functional tf:
	$(MAKE) ctu
	go test ./tests; status=$$?; $(MAKE) dtu; exit $$status

create-test-users ctu:
	docker compose exec -T postgres psql -U postgres -d user-service-db < tests/suite/create_test_users.sql

delete-test-users dtu:
	docker compose exec -T postgres psql -U postgres -d user-service-db < tests/suite/delete_test_users.sql

clear-db:
	docker compose exec -T postgres psql -U postgres -d user-service-db < tests/suite/clear_db.sql

up:
	docker compose -f docker-compose.yml up -d

up-dev:
	docker compose up -d

up-build:
	docker compose -f docker-compose.yml up -d --build

up-dev-build:
	docker compose up -d --build

down:
	docker compose down

prepare:
	$(MAKE) infra-dev
	@echo "Waiting for database to be ready..."
	@until docker compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; do sleep 1; done
	$(MAKE) migrate

infra:
	docker compose -f docker-compose.yml up -d postgres

infra-dev:
	docker compose up -d postgres

run:
	go run ./cmd/user-service/main.go
