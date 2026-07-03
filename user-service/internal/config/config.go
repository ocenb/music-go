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
	ContentServiceURL     string         `env:"CONTENT_SERVICE_URL" env-required:"true" validate:"url"`
	InternalServiceSecret string         `env:"INTERNAL_SERVICE_SECRET" env-required:"true"`
	DBConnectTimeout      time.Duration  `yaml:"db_connect_timeout" env:"DB_CONNECT_TIMEOUT" env-default:"10s" validate:"min=1s"`
	ShutdownTimeout       time.Duration  `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" env-default:"10s" validate:"min=1s"`
	Log                   LogConfig      `yaml:",inline"`
	Postgres              PostgresConfig `yaml:",inline"`
	GRPC                  GRPCConfig     `yaml:"grpc"`
	Auth                  AuthConfig     `yaml:",inline"`
	Kafka                 KafkaConfig    `yaml:",inline"`
}

type LogConfig struct {
	Level   int    `yaml:"log_level" env:"LOG_LEVEL" env-default:"0" validate:"oneof=-4 0 4 8"` // -4 = Debug, 0 = Info, 4 = Warn, 8 = Error
	Handler string `yaml:"log_handler" env:"LOG_HANDLER" env-default:"text" validate:"oneof=text json"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env:"GRPC_PORT" env-default:"9090" validate:"min=1,max=65535"`
	Timeout time.Duration `yaml:"timeout" env:"GRPC_TIMEOUT" env-default:"5s" validate:"min=100ms"`
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

type AuthConfig struct {
	JWTSecret            string        `env:"JWT_SECRET" env-required:"true"`
	BCryptCost           int           `env:"BCRYPT_COST" env-default:"12" validate:"min=4,max=31"`
	AccessTokenLiveTime  time.Duration `yaml:"access_token_live_time" env:"ACCESS_TOKEN_LIVE_TIME" env-default:"1h" validate:"min=1m"`
	RefreshTokenLiveTime time.Duration `yaml:"refresh_token_live_time" env:"REFRESH_TOKEN_LIVE_TIME" env-default:"720h" validate:"min=1h"`
	TokenCleanupInterval time.Duration `yaml:"token_cleanup_interval" env:"TOKEN_CLEANUP_INTERVAL" env-default:"24h" validate:"min=1m"`
}

type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS" env-required:"true"`
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
