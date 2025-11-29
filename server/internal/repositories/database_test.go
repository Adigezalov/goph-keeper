package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabaseRepository_InvalidDSN(t *testing.T) {
	// Test with clearly invalid DSN
	_, err := NewDatabaseRepository("invalid://dsn")
	assert.Error(t, err)
}

func TestDatabaseRepository_Struct(t *testing.T) {
	// Test that the struct can be created (without actual DB connection)
	dr := &DatabaseRepository{db: nil}
	assert.NotNil(t, dr)
}
