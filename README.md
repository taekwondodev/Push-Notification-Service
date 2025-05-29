# PushNotificationService


## Project Structure

/PushNotificationService/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handlers/
│   │   ├── websocket.go
│   │   └── notifications.go
│   ├── services/
│   │   ├── notification_service.go
│   │   └── kafka_service.go
│   ├── repositories/
│   │   └── notification_repository.go
│   ├── models/
│   │   └── notification.go
│   └── websocket/
│       └── hub.go
├── pkg/
│   └── logger/
│       └── logger.go
├── docker-compose.yml
├── go.mod
└── README.md
