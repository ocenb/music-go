package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Environment           string         `env:"ENVIRONMENT" env-default:"local" validate:"oneof=local dev test prod"`
	SearchServiceAddress  string         `env:"SEARCH_SERVICE_ADDRESS" env-required:"true"`
	UserServiceAddress    string         `env:"USER_SERVICE_ADDRESS" env-required:"true"`
	InternalServiceSecret string         `env:"INTERNAL_SERVICE_SECRET" env-required:"true"`
	Domain                string         `env:"DOMAIN" env-required:"true"`
	DBConnectTimeout      time.Duration  `yaml:"db_connect_timeout" env:"DB_CONNECT_TIMEOUT" env-default:"10s" validate:"min=1s"`
	ShutdownTimeout       time.Duration  `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" env-default:"30s" validate:"min=1s"`
	ImageFileLimit        int64          `yaml:"image_file_limit" env:"IMAGE_FILE_LIMIT" env-default:"10485760" validate:"min=1"`
	AudioFileLimit        int64          `yaml:"audio_file_limit" env:"AUDIO_FILE_LIMIT" env-default:"52428800" validate:"min=1"`
	Log                   LogConfig      `yaml:",inline"`
	Postgres              PostgresConfig `yaml:",inline"`
	HTTP                  HTTPConfig     `yaml:"http"`
	Cloudinary            CloudinaryConfig
	Kafka                 KafkaConfig `yaml:",inline"`
}

type LogConfig struct {
	Level   int    `yaml:"log_level" env:"LOG_LEVEL" env-default:"0" validate:"oneof=-4 0 4 8"`
	Handler string `yaml:"log_handler" env:"LOG_HANDLER" env-default:"text" validate:"oneof=text json"`
}

type HTTPConfig struct {
	Port         int           `yaml:"port" env:"PORT" env-default:"3000" validate:"min=1,max=65535"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"HTTP_READ_TIMEOUT" env-default:"60s" validate:"min=1s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-default:"60s" validate:"min=1s"`
}

type CloudinaryConfig struct {
	CloudName string `env:"CLOUDINARY_CLOUD_NAME" env-required:"true"`
	APIKey    string `env:"CLOUDINARY_API_KEY" env-required:"true"`
	APISecret string `env:"CLOUDINARY_API_SECRET" env-required:"true"`
}

type PostgresConfig struct {
	Host            string        `env:"POSTGRES_HOST" env-required:"true" validate:"hostname|ip"`
	Port            string        `env:"POSTGRES_PORT" env-required:"true" validate:"numeric"`
	User            string        `env:"POSTGRES_USER" env-required:"true"`
	Password        string        `env:"POSTGRES_PASSWORD" env-required:"true"`
	Name            string        `env:"POSTGRES_DB" env-required:"true"`
	SSLMode         string        `env:"POSTGRES_SSL_MODE" env-default:"disable" validate:"oneof=disable allow prefer require verify-ca verify-full"`
	MaxOpenConns    int32         `yaml:"db_max_open_conns" env:"POSTGRES_MAX_OPEN_CONNS" env-default:"10" validate:"min=1"`
	MaxIdleConns    int32         `yaml:"db_max_idle_conns" env:"POSTGRES_MAX_IDLE_CONNS" env-default:"5" validate:"min=1,ltefield=MaxOpenConns"`
	ConnMaxLifetime time.Duration `yaml:"db_conn_max_lifetime" env:"POSTGRES_CONN_MAX_LIFETIME" env-default:"1h" validate:"min=1m"`
	DSN             string
}

type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS" env-required:"true"`
	Topic   string   `env:"KAFKA_TOPIC" env-default:"email-notifications" validate:"required"`
}

func (p *PostgresConfig) buildDSN() {
	hostPort := net.JoinHostPort(p.Host, p.Port)
	p.DSN = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		p.User, p.Password, hostPort, p.Name, p.SSLMode)
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

	cfg.Postgres.buildDSN()

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
