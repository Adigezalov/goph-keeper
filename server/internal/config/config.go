package config

import (
	"flag"
	"os"
	"time"
)

const (
	DefaultServerAddress       = ":8080"
	DefaultDatabaseURI         = "postgres://user:password@localhost:5432/keeper?sslmode=disable"
	DefaultJWTSecret           = "your-secret-key-change-in-production"
	DefaultAccessTokenTTL      = 10 * time.Minute
	DefaultRefreshTokenTTL     = 120 * time.Hour
	DefaultSMTPHost            = "smtp.yandex.ru"
	DefaultSMTPPort            = "465"
	DefaultVerificationCodeTTL = 10 * time.Minute
)

type Config struct {
	ServerAddress       string
	DatabaseURI         string
	JWTSecret           string
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	SMTPHost            string
	SMTPPort            string
	SMTPUsername        string
	SMTPPassword        string
	SMTPFrom            string
	VerificationCodeTTL time.Duration
	TLSCertFile         string
	TLSKeyFile          string
}

func NewConfig() *Config {
	cfg := &Config{}

	serverAddress := DefaultServerAddress
	databaseURI := DefaultDatabaseURI
	var jwtSecret string
	smtpHost := DefaultSMTPHost
	smtpPort := DefaultSMTPPort
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")
	var tlsCertFile string
	var tlsKeyFile string

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		serverAddress = envRunAddr
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		databaseURI = envDatabaseURI
	}
	if envJWTSecret := os.Getenv("JWT_SECRET"); envJWTSecret != "" {
		jwtSecret = envJWTSecret
	}
	if envSMTPHost := os.Getenv("SMTP_HOST"); envSMTPHost != "" {
		smtpHost = envSMTPHost
	}
	if envSMTPPort := os.Getenv("SMTP_PORT"); envSMTPPort != "" {
		smtpPort = envSMTPPort
	}
	if envTLSCertFile := os.Getenv("TLS_CERT_FILE"); envTLSCertFile != "" {
		tlsCertFile = envTLSCertFile
	}
	if envTLSKeyFile := os.Getenv("TLS_KEY_FILE"); envTLSKeyFile != "" {
		tlsKeyFile = envTLSKeyFile
	}

	flag.StringVar(&cfg.ServerAddress, "a", serverAddress, "адрес и порт запуска сервиса")
	flag.StringVar(&cfg.DatabaseURI, "d", databaseURI, "адрес подключения к базе данных")
	flag.StringVar(&cfg.JWTSecret, "jwt-secret", jwtSecret, "секретный ключ для JWT токенов (обязательный)")
	flag.StringVar(&cfg.SMTPHost, "smtp-host", smtpHost, "SMTP хост для отправки email")
	flag.StringVar(&cfg.SMTPPort, "smtp-port", smtpPort, "SMTP порт")
	flag.StringVar(&cfg.TLSCertFile, "tls-cert", tlsCertFile, "путь к TLS сертификату (обязательный)")
	flag.StringVar(&cfg.TLSKeyFile, "tls-key", tlsKeyFile, "путь к TLS приватному ключу (обязательный)")

	cfg.SMTPUsername = smtpUsername
	cfg.SMTPPassword = smtpPassword
	cfg.SMTPFrom = smtpFrom
	cfg.AccessTokenTTL = DefaultAccessTokenTTL
	cfg.RefreshTokenTTL = DefaultRefreshTokenTTL
	cfg.VerificationCodeTTL = DefaultVerificationCodeTTL

	flag.Parse()

	cfg.normalize()
	cfg.validate()

	return cfg
}

func (c *Config) normalize() {
	if len(c.ServerAddress) == 0 {
		c.ServerAddress = DefaultServerAddress
		return
	}
	if c.ServerAddress[0] != ':' {
		if c.ServerAddress[0] >= '0' && c.ServerAddress[0] <= '9' {
			c.ServerAddress = ":" + c.ServerAddress
		}
	}
}

func (c *Config) validate() {
	if c.ServerAddress == "" {
		panic("ServerAddress cannot be empty")
	}
	if c.JWTSecret == "" {
		panic("JWT_SECRET must be set via environment variable or -jwt-secret flag. Using default value is a security risk.")
	}
	if c.TLSCertFile == "" {
		panic("TLS_CERT_FILE must be set via environment variable or -tls-cert flag. TLS is required for security.")
	}
	if c.TLSKeyFile == "" {
		panic("TLS_KEY_FILE must be set via environment variable or -tls-key flag. TLS is required for security.")
	}
}
