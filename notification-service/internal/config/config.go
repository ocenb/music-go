package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Environment     string        `env:"ENVIRONMENT" env-default:"local" validate:"oneof=local dev test prod"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" env-default:"10s" validate:"min=1s"`
	Log             LogConfig     `yaml:",inline"`
	Kafka           KafkaConfig   `yaml:",inline"`
	SMTP            SMTPConfig    `yaml:",inline"`
}

type LogConfig struct {
	Level   int    `yaml:"log_level" env:"LOG_LEVEL" env-default:"0" validate:"oneof=-4 0 4 8"`
	Handler string `yaml:"log_handler" env:"LOG_HANDLER" env-default:"text" validate:"oneof=text json"`
}

type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS" env-required:"true"`
	Topic   string   `env:"KAFKA_TOPIC" env-required:"true"`
	GroupID string   `env:"KAFKA_GROUP_ID" env-required:"true"`
}

type SMTPConfig struct {
	Host     string `env:"SMTP_HOST" env-required:"true" validate:"hostname|ip"`
	Port     int    `env:"SMTP_PORT" env-required:"true" validate:"min=1,max=65535"`
	Username string `env:"SMTP_USERNAME" env-required:"true"`
	Password string `env:"SMTP_PASSWORD" env-required:"true"`
}

func Load() (*Config, error) {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local"
	}

	switch env {
	case "local", "dev", "test", "prod":
	default:
		env = "local"
	}

	configPath := filepath.Join("config", env+".yaml")
	var cfg Config

	//nolint:gosec // configPath is built from an allowlisted ENVIRONMENT value
	if _, err := os.Stat(configPath); err == nil {
		if err := loadAndValidate(configPath, &cfg); err != nil {
			return nil, err
		}
	} else {
		if err := loadAndValidateOnlyEnv(&cfg); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}

func loadAndValidate(configPath string, cfg any) error {
	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		return fmt.Errorf("cannot read config from file %s: %w", configPath, err)
	}
	return validateStruct(cfg)
}

func loadAndValidateOnlyEnv(cfg any) error {
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return fmt.Errorf("cannot read config from environment: %w", err)
	}
	return validateStruct(cfg)
}

func validateStruct(cfg any) error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	return nil
}
