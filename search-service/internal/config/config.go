package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/ocenb/music-go/search-service/internal/utils"
)

type Config struct {
	GRPC               GRPCConfig `yaml:"grpc"`
	LogLevel           int        `yaml:"log_level" env-default:"0"`
	LogHandler         string     `yaml:"log_handler" env-default:"text"`
	Environment        string     `env:"ENVIRONMENT" env-required:"true"`
	ElasticHost        string     `env:"ELASTIC_HOST" env-required:"true"`
	ElasticPort        string     `env:"ELASTIC_PORT" env-required:"true"`
	ElasticUser        string     `env:"ELASTIC_USER" env-required:"true"`
	ElasticPassword    string     `env:"ELASTIC_PASSWORD" env-required:"true"`
	ElasticUrl         string
	UserServiceAddress string `env:"USER_SERVICE_ADDRESS" env-required:"true"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-default:"9090"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
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

	cfg.ElasticUrl = utils.GetElasticUrl(cfg.ElasticHost, cfg.ElasticPort, cfg.ElasticUser, cfg.ElasticPassword)

	return &cfg
}
