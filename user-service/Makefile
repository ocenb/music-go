ENV_TARGETS := run migrate prepare
ENV_TEST_TARGETS := tf test-functional

ifneq ($(filter $(MAKECMDGOALS),$(ENV_TARGETS)),)
ifeq ($(wildcard .env),)
$(shell echo ".env not found. Copying from .env.example..." && cp .env.example .env)
endif
include .env
export
endif

ifneq ($(filter $(MAKECMDGOALS),$(ENV_TEST_TARGETS)),)
ifeq ($(wildcard tests/.env.test),)
$(shell echo "tests/.env.test not found. Copying from tests/.env.test.example..." && cp tests/.env.test.example tests/.env.test)
endif
include tests/.env.test
export
endif

DB_URL := postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

.PHONY: build fix lint migrate gen-mocks test-unit tu test-functional tf \
        run up up-dev up-build up-dev-build down prepare infra infra-dev logs test-cover

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
	go test ./internal/... | grep -v '\[no test files\]'

test-cover:
	go test -cover ./internal/...

test-functional tf:
	go test ./tests/... -count=1 -timeout 5m

up:
	docker compose -f compose.yaml up -d

up-dev:
	docker compose up -d

up-build:
	docker compose -f compose.yaml up -d --build

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
	docker compose -f compose.yaml up -d postgres

infra-dev:
	docker compose up -d postgres

run:
	go run ./cmd/user-service/main.go
