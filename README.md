# Music-Go

Проект Music-Go - это микросервисная архитектура для музыкальной платформы, разработанная на языке Go. Система состоит из четырех основных микросервисов, каждый из которых отвечает за свой функциональный домен.

## Архитектура проекта

![Архитектура](assets/image.png)

### Stack

- Go
- PostgreSQL, Elasticsearch
- Kafka, gRPC, HTTP(REST)
- Gin, Buf, Ffmpeg
- Docker, Nginx, Kubernetes, Kustomize
- Cloudinary (file storage)

### Frontend

<https://github.com/ocenb/music-app-frontend>

### Proto файлы

<https://github.com/ocenb/music-protos>

### User Service

Сервис аутентификации и авторизации пользователей. Управляет регистрацией, авторизацией, профилями пользователей и токенами доступа.

**Технологии:**

- PostgreSQL для хранения данных пользователей
- gRPC для коммуникации с другими сервисами
- Kafka для отправки уведомлений
- Аутентификация с использованием JWT токенов (access и refresh)

### Content Service

Сервис для управления музыкальным контентом (треки, плейлисты, история прослушивания).

**Технологии:**

- PostgreSQL для хранения данных
- HTTP REST API на основе Gin для обработки запросов
- gRPC-клиенты к User Service и Search Service
- Kafka для отправки уведомлений
- Cloudinary для хранения аудио и обложек
- FFmpeg для обработки аудиофайлов

### Search Service

Сервис для полнотекстового поиска по музыкальному контенту и пользователям.

**Технологии:**

- Elasticsearch для индексирования и поиска
- gRPC для коммуникации с другими сервисами

### Notification Service

Сервис для обработки и отправки уведомлений пользователям.

**Технологии:**

- Kafka для приема сообщений от других сервисов
- SMTP для отправки email уведомлений

## Тестирование

Команды выполняются из каталога нужного сервиса:

- **Юнит-тесты**

```bash
cd user-service   # или content-service, search-service, notification-service
make tu
```

- **Функциональные тесты**

```bash
cd user-service
make tf
```

Для unit-тестирования используются моки, генерируемые mockery (`make gen-mocks`).

## Запуск проекта

### Требования

- Docker и Docker Compose
- Go 1.26+
- Make

### Общая сеть Docker

Создайте общую сеть для микросервисов:

```bash
docker network create music-go-network
```

### Запуск сервисов

1. User Service:

```bash
cd user-service
make up-dev
```

2. Content Service:

```bash
cd content-service
make up-dev
```

3. Search Service:

```bash
cd search-service
make up-dev
```

4. Notification Service:

```bash
cd notification-service
make up-dev
```

### API Gateway

После запуска сервисов поднимите Nginx gateway из корня репозитория — он проксирует HTTP-запросы к Content Service и gRPC к User Service:

```bash
docker compose up -d
```

### Kubernetes

Манифесты в `k8s/` (Kustomize): `base/` + `overlays/local` и `overlays/prod`.

```bash
# images
docker build -t music-go/user-service:latest ./user-service
docker build -t music-go/content-service:latest ./content-service
docker build -t music-go/search-service:latest ./search-service
docker build -t music-go/notification-service:latest ./notification-service

# ingress
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.11.3/deploy/static/provider/cloud/deploy.yaml

# deploy
kubectl apply -k k8s/overlays/local
```

## Структура кода

Все сервисы следуют схожей структуре:

```
service-name/
├── cmd/
│   └── service-name/       # Точка входа (main.go)
├── internal/
│   ├── config/             # Загрузка конфигурации из env и YAML
│   ├── server/             # Сборка и запуск HTTP/gRPC сервера
│   ├── handlers/           # HTTP/gRPC обработчики
│   ├── middlewares/        # Auth, logging (user, content, search)
│   ├── services/           # Бизнес-логика (пакеты по доменам)
│   ├── repos/              # Доступ к данным
│   ├── models/             # Модели и DTO
│   ├── clients/            # Клиенты других сервисов и внешних API
│   ├── storage/            # Подключения к БД, Kafka, Elasticsearch
│   ├── errs/               # Типизированные ошибки, маппинг в HTTP/gRPC
│   └── logger/             # Structured logging (slog)
├── config/
│   └── local.yaml          # Локальные настройки по умолчанию
├── migrations/             # SQL-миграции (user, content), goose
├── tests/                  # E2E-тесты
├── Dockerfile
├── compose.yaml
├── compose.override.yaml   # Локальные переопределения Docker Compose
├── .env.example
└── Makefile                # build, migrate, tu, tf и др.
```

Отличия по сервисам:

- **user-service, content-service** — PostgreSQL (`storage/postgres`, `storage/migrator`, `storage/transactor`), SQL-миграции в `migrations/`
- **search-service** — Elasticsearch (`storage/elastic`)
- **notification-service** — Kafka (`storage/kafka`, `internal/kafka`), SMTP-клиент (`clients/smtp`)

## Коммуникация между сервисами

| Сервис | Принимает | Отправляет |
|--------|-----------|------------|
| **user-service** | gRPC (регистрация, auth, профили) | gRPC → search, content; Kafka → notification |
| **content-service** | HTTP REST (треки, плейлисты, история) | gRPC → user, search; Kafka → notification; Cloudinary |
| **search-service** | gRPC (поиск) | gRPC → user; Elasticsearch |
| **notification-service** | Kafka (события) | SMTP (email) |

Nginx gateway в корне проекта маршрутизирует внешние запросы: `/api/content/*` → content-service, gRPC → user-service.
