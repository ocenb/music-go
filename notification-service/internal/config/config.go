package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	LogLevel     int      `yaml:"log_level" env-default:"0"`
	LogHandler   string   `yaml:"log_handler" env-default:"text"`
	Environment  string   `env:"ENVIRONMENT" env-required:"true"`
	KafkaBrokers []string `env:"KAFKA_BROKERS" env-required:"true"`
	KafkaTopic   string   `env:"KAFKA_TOPIC" env-required:"true"`
	KafkaGroupID string   `env:"KAFKA_GROUP_ID" env-required:"true"`
	SMTPHost     string   `env:"SMTP_HOST" env-required:"true"`
	SMTPPort     int      `env:"SMTP_PORT" env-required:"true"`
	SMTPUsername string   `env:"SMTP_USERNAME" env-required:"true"`
	SMTPPassword string   `env:"SMTP_PASSWORD" env-required:"true"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	configPath := ""
	if os.Getenv("ENVIRONMENT") == "local" {
		configPath = "config/local.yaml"
	} else if os.Getenv("ENVIRONMENT") == "prod" {
		configPath = "config/prod.yaml"
	} else {
		log.Fatalf("Invalid environment: %s", os.Getenv("ENVIRONMENT"))
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file not found: %s", configPath)
	}

	var cfg Config

	err = cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	return &cfg
}
