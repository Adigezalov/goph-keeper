package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Normalize_WithPort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Port only",
			input:    ":8080",
			expected: ":8080",
		},
		{
			name:     "Port number without colon",
			input:    "8080",
			expected: ":8080",
		},
		{
			name:     "Port number with leading digit",
			input:    "3000",
			expected: ":3000",
		},
		{
			name:     "Hostname with port",
			input:    "localhost:8080",
			expected: "localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				ServerAddress: tt.input,
			}
			cfg.normalize()
			assert.Equal(t, tt.expected, cfg.ServerAddress)
		})
	}
}

func TestConfig_Normalize_Various(t *testing.T) {
	t.Run("Already has colon", func(t *testing.T) {
		cfg := &Config{ServerAddress: ":9090"}
		cfg.normalize()
		assert.Equal(t, ":9090", cfg.ServerAddress)
	})

	t.Run("Numeric without colon", func(t *testing.T) {
		cfg := &Config{ServerAddress: "7070"}
		cfg.normalize()
		assert.Equal(t, ":7070", cfg.ServerAddress)
	})

	t.Run("Host and port", func(t *testing.T) {
		cfg := &Config{ServerAddress: "example.com:8080"}
		cfg.normalize()
		assert.Equal(t, "example.com:8080", cfg.ServerAddress)
	})
}

func TestDefaultValues(t *testing.T) {
	assert.Equal(t, ":8080", DefaultServerAddress)
	assert.Equal(t, "postgres://user:password@localhost:5432/keeper?sslmode=disable", DefaultDatabaseURI)
	assert.Equal(t, "your-secret-key-change-in-production", DefaultJWTSecret)
	assert.Equal(t, 10*time.Minute, DefaultAccessTokenTTL)
	assert.Equal(t, 120*time.Hour, DefaultRefreshTokenTTL)
}

func TestConfig_Struct(t *testing.T) {
	cfg := &Config{
		ServerAddress:   ":8080",
		DatabaseURI:     "postgres://localhost:5432/test",
		JWTSecret:       "test-secret",
		AccessTokenTTL:  DefaultAccessTokenTTL,
		RefreshTokenTTL: DefaultRefreshTokenTTL,
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, ":8080", cfg.ServerAddress)
	assert.Equal(t, DefaultAccessTokenTTL, cfg.AccessTokenTTL)
	assert.Equal(t, DefaultRefreshTokenTTL, cfg.RefreshTokenTTL)
}

func TestConfig_AllFields(t *testing.T) {
	cfg := &Config{
		ServerAddress:   ":9090",
		DatabaseURI:     "postgres://user:pass@host:5432/db",
		JWTSecret:       "my-secret-key",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
	}

	assert.Equal(t, ":9090", cfg.ServerAddress)
	assert.Equal(t, "postgres://user:pass@host:5432/db", cfg.DatabaseURI)
	assert.Equal(t, "my-secret-key", cfg.JWTSecret)
	assert.Equal(t, 15*time.Minute, cfg.AccessTokenTTL)
	assert.Equal(t, 24*time.Hour, cfg.RefreshTokenTTL)
}

func TestConfig_Normalize_ComplexCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Port 80",
			input:    "80",
			expected: ":80",
		},
		{
			name:     "Port 443",
			input:    "443",
			expected: ":443",
		},
		{
			name:     "Port 3000",
			input:    "3000",
			expected: ":3000",
		},
		{
			name:     "Already normalized :3000",
			input:    ":3000",
			expected: ":3000",
		},
		{
			name:     "With hostname",
			input:    "app.example.com:8080",
			expected: "app.example.com:8080",
		},
		{
			name:     "localhost with port",
			input:    "localhost:9090",
			expected: "localhost:9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{ServerAddress: tt.input}
			cfg.normalize()
			assert.Equal(t, tt.expected, cfg.ServerAddress)
		})
	}
}

func TestDefaultConstants(t *testing.T) {
	t.Run("DefaultServerAddress", func(t *testing.T) {
		assert.Equal(t, ":8080", DefaultServerAddress)
	})

	t.Run("DefaultDatabaseURI", func(t *testing.T) {
		assert.Contains(t, DefaultDatabaseURI, "postgres://")
		assert.Contains(t, DefaultDatabaseURI, "keeper")
	})

	t.Run("DefaultJWTSecret", func(t *testing.T) {
		assert.NotEmpty(t, DefaultJWTSecret)
	})

	t.Run("DefaultAccessTokenTTL", func(t *testing.T) {
		assert.Equal(t, 10*time.Minute, DefaultAccessTokenTTL)
	})

	t.Run("DefaultRefreshTokenTTL", func(t *testing.T) {
		assert.Equal(t, 120*time.Hour, DefaultRefreshTokenTTL)
	})
}

func TestConfig_TTLValues(t *testing.T) {
	cfg := &Config{
		AccessTokenTTL:  DefaultAccessTokenTTL,
		RefreshTokenTTL: DefaultRefreshTokenTTL,
	}

	assert.Equal(t, 10*time.Minute, cfg.AccessTokenTTL)
	assert.Equal(t, 120*time.Hour, cfg.RefreshTokenTTL)

	// Verify the values in seconds
	assert.Equal(t, int64(600), int64(cfg.AccessTokenTTL.Seconds()))
	assert.Equal(t, int64(432000), int64(cfg.RefreshTokenTTL.Seconds()))
}

func TestConfig_Normalize_EdgeCases(t *testing.T) {
	t.Run("Single character numeric", func(t *testing.T) {
		cfg := &Config{ServerAddress: "8"}
		cfg.normalize()
		assert.Equal(t, ":8", cfg.ServerAddress)
	})

	t.Run("Starts with letter", func(t *testing.T) {
		cfg := &Config{ServerAddress: "localhost"}
		cfg.normalize()
		assert.Equal(t, "localhost", cfg.ServerAddress)
	})

	t.Run("IPv4 address", func(t *testing.T) {
		cfg := &Config{ServerAddress: "127.0.0.1:8080"}
		cfg.normalize()
		// normalize adds ':' prefix to addresses starting with digit
		assert.Equal(t, ":127.0.0.1:8080", cfg.ServerAddress)
	})

	t.Run("Numeric IP", func(t *testing.T) {
		cfg := &Config{ServerAddress: "192.168.1.1:3000"}
		cfg.normalize()
		// normalize adds ':' prefix to addresses starting with digit
		assert.Equal(t, ":192.168.1.1:3000", cfg.ServerAddress)
	})
}
