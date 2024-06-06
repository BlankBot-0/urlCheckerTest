package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env     string `yaml:"env" env-default:"local"`
	Checker `yaml:"service"`
}

type Checker struct {
	URLs        []string `yaml:"urls" env-required:"true"`
	RateLimiter `yaml:"rate_limiter"`
}

type RateLimiter struct {
	RateLimit          time.Duration `yaml:"rate_limit" env-default:"15s"`
	BackoffCoefficient int           `yaml:"backoff_coefficient" env-default:"2"`
	MaxDelay           time.Duration `yaml:"max_delay" env-default:"15m"`
}

func MustLoad() *Config {
	envFile, err := godotenv.Read(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	configPath := envFile["CONFIG_PATH"]
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable not set")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("failed to read config: %s", err)
	}
	return &cfg
}