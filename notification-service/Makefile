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

.PHONY: build fix lint test-unit tu test-functional tf \
        run up up-dev up-build up-dev-build down prepare infra infra-dev test-cover

build:
	go build -o .build/notification-service ./cmd/notification-service/main.go

fix:
	golangci-lint run --fix

lint:
	golangci-lint run

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
	@echo "Waiting for kafka to be ready..."
	@until docker compose exec -T kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list >/dev/null 2>&1; do sleep 1; done

infra:
	docker compose -f compose.yaml up -d kafka zookeeper

infra-dev:
	docker compose up -d kafka zookeeper

run:
	go run ./cmd/notification-service/main.go
