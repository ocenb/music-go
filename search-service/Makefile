ENV_TARGETS := run prepare
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

.PHONY: build fix lint gen-mocks test-unit tu test-functional tf \
        run up up-dev up-build up-dev-build down prepare infra infra-dev test-cover

build:
	go build -o .build/search-service ./cmd/search-service/main.go

fix:
	golangci-lint run --fix

lint:
	golangci-lint run

gen-mocks:
	mockery

test-unit tu:
	go test ./internal/...

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
	@echo "Waiting for elasticsearch to be ready..."
	@until docker compose exec -T elasticsearch curl -sf http://localhost:9200 >/dev/null 2>&1; do sleep 1; done

infra:
	docker compose -f compose.yaml up -d elasticsearch

infra-dev:
	docker compose up -d elasticsearch

run:
	go run ./cmd/search-service/main.go
