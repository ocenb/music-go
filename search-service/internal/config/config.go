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
	Environment           string        `env:"ENVIRONMENT" env-default:"local" validate:"oneof=local dev test prod"`
	UserServiceAddress    string        `env:"USER_SERVICE_ADDRESS" env-required:"true"`
	ElasticConnectTimeout time.Duration `yaml:"elastic_connect_timeout" env:"ELASTIC_CONNECT_TIMEOUT" env-default:"10s" validate:"min=1s"`
	ShutdownTimeout       time.Duration `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" env-default:"10s" validate:"min=1s"`
	Log                   LogConfig     `yaml:",inline"`
	Elastic               ElasticConfig `yaml:",inline"`
	GRPC                  GRPCConfig    `yaml:"grpc"`
}

type LogConfig struct {
	Level   int    `yaml:"log_level" env:"LOG_LEVEL" env-default:"0" validate:"oneof=-4 0 4 8"`
	Handler string `yaml:"log_handler" env:"LOG_HANDLER" env-default:"text" validate:"oneof=text json"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env:"GRPC_PORT" env-default:"9091" validate:"min=1,max=65535"`
	Timeout time.Duration `yaml:"timeout" env:"GRPC_TIMEOUT" env-default:"5s" validate:"min=100ms"`
}

type ElasticConfig struct {
	Host     string `env:"ELASTIC_HOST" env-required:"true" validate:"hostname|ip"`
	Port     string `env:"ELASTIC_PORT" env-required:"true" validate:"numeric"`
	User     string `env:"ELASTIC_USER" env-default:"elastic"`
	Password string `env:"ELASTIC_PASSWORD" env-default:""`
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
