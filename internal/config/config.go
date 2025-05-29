package config

import (
	"os"
)

type Config struct {
    Server ServerConfig
    Mongo  MongoConfig
    Kafka  KafkaConfig
}

type ServerConfig struct {
    Port string
}

type MongoConfig struct {
    URI      string
    Database string
}

type KafkaConfig struct {
    Brokers []string
    Topic   string
    GroupID string
}

func Load() *Config {
    return &Config{
        Server: ServerConfig{
            Port: getEnv("PORT", "8080"),
        },
        Mongo: MongoConfig{
            URI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
            Database: getEnv("MONGO_DATABASE", "notificationsdb"),
        },
        Kafka: KafkaConfig{
            Brokers: []string{getEnv("KAFKA_BROKER", "localhost:9092")},
            Topic:   getEnv("KAFKA_TOPIC", "notifications"),
            GroupID: getEnv("KAFKA_GROUP_ID", "websocket-notifier"),
        },
    }
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}