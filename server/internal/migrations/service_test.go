package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	service := NewService(nil, "/path/to/migrations")
	assert.NotNil(t, service)
	assert.Equal(t, "/path/to/migrations", service.migrationsPath)
}

func TestMigrationStruct(t *testing.T) {
	migration := Migration{
		Version: 1,
		Name:    "create_users_table",
		UpSQL:   "CREATE TABLE users (...)",
		DownSQL: "DROP TABLE users",
	}

	assert.Equal(t, 1, migration.Version)
	assert.Equal(t, "create_users_table", migration.Name)
	assert.NotEmpty(t, migration.UpSQL)
	assert.NotEmpty(t, migration.DownSQL)
}

func TestMigrationStatusStruct(t *testing.T) {
	appliedAt := "2024-01-01 12:00:00"
	status := MigrationStatus{
		Version:   1,
		Name:      "create_users_table",
		Applied:   true,
		AppliedAt: &appliedAt,
	}

	assert.Equal(t, 1, status.Version)
	assert.True(t, status.Applied)
	assert.NotNil(t, status.AppliedAt)
}

func TestMigrationStatusStruct_NotApplied(t *testing.T) {
	status := MigrationStatus{
		Version:   2,
		Name:      "add_index",
		Applied:   false,
		AppliedAt: nil,
	}

	assert.Equal(t, 2, status.Version)
	assert.False(t, status.Applied)
	assert.Nil(t, status.AppliedAt)
}
