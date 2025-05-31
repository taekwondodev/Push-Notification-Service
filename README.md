# Push Notification Service

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.24.3-blue.svg)
![Docker](https://img.shields.io/badge/docker-ready-blue.svg)
![Kafka](https://img.shields.io/badge/kafka-3.7-orange.svg)
![MongoDB](https://img.shields.io/badge/mongodb-6.0-green.svg)
![WebSocket](https://img.shields.io/badge/websocket-ready-purple.svg)

A real-time push notification service built with Go, featuring WebSocket connections, Kafka message queuing, and MongoDB persistence. Designed for microservices architecture with JWT authentication support.

</div>

## Features

- **Real-time notifications** via WebSocket connections
- **Async message processing** with Apache Kafka
- **Persistent storage** with MongoDB
- **JWT Authentication** ready (gateway integration)
- **Docker containerized** for easy deployment
- **RESTful API** for notification management
- **Graceful shutdown** support
- **Production ready** with proper logging and error handling


## Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Gateway   │───▶│ Notification│───▶│  WebSocket  │
│ (Auth/JWT)  │    │   Service   │    │   Clients   │
└─────────────┘    └─────────────┘    └─────────────┘
                           │
                           ▼
                   ┌─────────────┐    ┌─────────────┐
                   │    Kafka    │───▶│   MongoDB   │
                   │   Queue     │    │  Database   │
                   └─────────────┘    └─────────────┘
```

## API Endpoints

Complete API documentation is available in OpenAPI 3.0 format:

- [![Open in Swagger Editor](https://img.shields.io/badge/Swagger-Editor-%23Clojure?style=for-the-badge&logo=swagger)](https://editor.swagger.io/?url=https://raw.githubusercontent.com/taekwondodev/push-notification-service/main/api/openapi.yaml)

- [Raw OpenAPI Spec](api/openapi.yaml)

## Authentication

The service expects JWT validation to be handled by an upstream gateway. The gateway should:

1. Validate JWT tokens
2. Extract user information
3. Forward requests with `X-User-Username` header

For development, you can test without authentication by setting the header manually.

## Quick Start

### Prerequisites

- Go 1.23+
- Docker

### Using Docker Compose

```bash
git clone https://github.com/taekwondodev/push-notification-service.git
cd push-notification-service

docker-compose up
```

## Configuration

The service is configured via environment variables:

| Variable         | Default                     | Description               |
|------------------|-----------------------------|---------------------------|
| `PORT`           | `8080`                      | HTTP server port          |
| `MONGO_URI`      | `mongodb://localhost:27017` | MongoDB connection string |
| `MONGO_DATABASE` | `notificationsdb`           | MongoDB database name     |
| `KAFKA_BROKER`   | `localhost:9092`            | Kafka broker address      |
| `KAFKA_TOPIC`    | `notifications`             | Kafka topic name          |
| `KAFKA_GROUP_ID` | `websocket-notifier`        | Kafka consumer group ID   |

## Project Structure

```
.
├── cmd/server/           # Application entry point
├── internal/
│   ├── api/              # HTTP server and routing
│   ├── config/           # Configuration management
│   ├── controller/       # HTTP handlers
│   ├── middleware/       # HTTP middleware
│   ├── models/           # Data models
│   ├── repository/       # Data access layer
│   ├── service/          # Business logic
│   └── websocket/        # WebSocket hub
├── client-test/          # Test HTML client
├── docker-compose.yml    # Docker services
├── Dockerfile            # Go service container
└── Dockerfile.test       # Test client container
```

## Testing

```bash
docker build -f Dockerfile.test -t notification-client . && \
docker run --rm -p 3000:80 notification-client

open http://localhost:3000/
```