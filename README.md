# PushNotificationService


## Project Structure

/PushNotificationService/
├── docker-compose.yml
├── go.mod
├── internal/
│   ├── kafka/
│   │   ├── producer.go
│   │   └── consumer.go
│   ├── websocket/
│   │   └── server.go
│   ├── db/
│   │   └── mongo.go
│   └── models/
│       └── notification.go
├── cmd/
│   └── server/main.go
├── web/
│   └── client.html
└── README.md
