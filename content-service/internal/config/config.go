package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/ocenb/music-go/content-service/internal/utils"
)

type Config struct {
	LogLevel             int           `yaml:"log_level" env-default:"0"`
	LogHandler           string        `yaml:"log_handler" env-default:"text"`
	DBMaxOpenConns       int           `yaml:"db_max_open_conns" env-default:"10"`
	DBMaxIdleConns       int           `yaml:"db_max_idle_conns" env-default:"5"`
	DBConnMaxLifetime    time.Duration `yaml:"db_conn_max_lifetime" env-default:"1h"`
	ImageFileLimit       int64         `yaml:"image_file_limit" env-default:"10485760"`
	AudioFileLimit       int64         `yaml:"audio_file_limit" env-default:"52428800"`
	Environment          string        `env:"ENVIRONMENT" env-required:"true"`
	DBHost               string        `env:"POSTGRES_HOST" env-required:"true"`
	DBPort               string        `env:"POSTGRES_PORT" env-required:"true"`
	DBUser               string        `env:"POSTGRES_USER" env-required:"true"`
	DBPassword           string        `env:"POSTGRES_PASSWORD" env-required:"true"`
	DBName               string        `env:"POSTGRES_DB" env-required:"true"`
	DBSSLMode            string        `env:"POSTGRES_SSL_MODE" env-default:"disable"`
	DatabaseUrl          string
	RedisHost            string `env:"REDIS_HOST" env-required:"true"`
	RedisPort            string `env:"REDIS_PORT" env-required:"true"`
	RedisPassword        string `env:"REDIS_PASSWORD" env-required:"true"`
	RedisUrl             string
	Domain               string   `env:"DOMAIN" env-required:"true"`
	Port                 int      `env:"PORT" env-required:"true"`
	CloudinaryCloudName  string   `env:"CLOUDINARY_CLOUD_NAME" env-required:"true"`
	CloudinaryApiKey     string   `env:"CLOUDINARY_API_KEY" env-required:"true"`
	CloudinaryApiSecret  string   `env:"CLOUDINARY_API_SECRET" env-required:"true"`
	SearchServiceAddress string   `env:"SEARCH_SERVICE_ADDRESS" env-required:"true"`
	UserServiceAddress   string   `env:"USER_SERVICE_ADDRESS" env-required:"true"`
	KafkaBrokers         []string `env:"KAFKA_BROKERS" env-required:"true"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found: %v", err)
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

	cfg.DatabaseUrl = utils.GetPostgresUrl(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	cfg.RedisUrl = utils.GetRedisUrl(cfg.RedisHost, cfg.RedisPort)
	return &cfg
}
