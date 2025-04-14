package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/ocenb/music-go/user-service/internal/utils"
)

type Config struct {
	GRPC                 GRPCConfig    `yaml:"grpc"`
	LogLevel             int           `yaml:"log_level" env-default:"0"`
	LogHandler           string        `yaml:"log_handler" env-default:"text"`
	AccessTokenLiveTime  time.Duration `yaml:"access_token_live_time" env-default:"1h"`
	RefreshTokenLiveTime time.Duration `yaml:"refresh_token_live_time" env-default:"720h"`
	DBMaxOpenConns       int           `yaml:"db_max_open_conns" env-default:"10"`
	DBMaxIdleConns       int           `yaml:"db_max_idle_conns" env-default:"5"`
	DBConnMaxLifetime    time.Duration `yaml:"db_conn_max_lifetime" env-default:"1h"`
	Environment          string        `env:"ENVIRONMENT" env-required:"true"`
	JWTSecret            string        `env:"JWT_SECRET" env-required:"true"`
	BCryptCost           int           `env:"BCRYPT_COST" env-default:"12"`
	DBHost               string        `env:"POSTGRES_HOST" env-required:"true"`
	DBPort               string        `env:"POSTGRES_PORT" env-required:"true"`
	DBUser               string        `env:"POSTGRES_USER" env-required:"true"`
	DBPassword           string        `env:"POSTGRES_PASSWORD" env-required:"true"`
	DBName               string        `env:"POSTGRES_DB" env-required:"true"`
	DBSSLMode            string        `env:"POSTGRES_SSL_MODE" env-default:"disable"`
	DatabaseUrl          string
	SearchServiceAddress string   `env:"SEARCH_SERVICE_ADDRESS" env-required:"true"`
	ContentServiceURL    string   `env:"CONTENT_SERVICE_URL" env-required:"true"`
	KafkaBrokers         []string `env:"KAFKA_BROKERS" env-required:"true"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-default:"9090"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
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

	cfg.DatabaseUrl = utils.GetDBUrl(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	return &cfg
}
