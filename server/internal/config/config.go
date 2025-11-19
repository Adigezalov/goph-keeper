package config

import (
	"flag"
	"os"
	"time"
)

const (
	DefaultServerAddress   = ":8080"
	DefaultDatabaseURI     = "postgres://user:password@localhost:5432/keeper?sslmode=disable"
	DefaultJWTSecret       = "your-secret-key-change-in-production"
	DefaultAccessTokenTTL  = 10 * time.Minute
	DefaultRefreshTokenTTL = 120 * time.Hour
)

type Config struct {
	ServerAddress   string
	DatabaseURI     string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func NewConfig() *Config {
	cfg := &Config{}

	serverAddress := DefaultServerAddress
	databaseURI := DefaultDatabaseURI
	jwtSecret := DefaultJWTSecret

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		serverAddress = envRunAddr
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		databaseURI = envDatabaseURI
	}
	if envJWTSecret := os.Getenv("JWT_SECRET"); envJWTSecret != "" {
		jwtSecret = envJWTSecret
	}

	flag.StringVar(&cfg.ServerAddress, "a", serverAddress, "адрес и порт запуска сервиса")
	flag.StringVar(&cfg.DatabaseURI, "d", databaseURI, "адрес подключения к базе данных")
	flag.StringVar(&cfg.JWTSecret, "jwt-secret", jwtSecret, "секретный ключ для JWT токенов")

	cfg.AccessTokenTTL = DefaultAccessTokenTTL
	cfg.RefreshTokenTTL = DefaultRefreshTokenTTL

	flag.Parse()

	cfg.normalize()

	return cfg
}

func (c *Config) normalize() {
	if c.ServerAddress[0] != ':' && len(c.ServerAddress) > 0 {
		if c.ServerAddress[0] >= '0' && c.ServerAddress[0] <= '9' {
			c.ServerAddress = ":" + c.ServerAddress
		}
	}
}
