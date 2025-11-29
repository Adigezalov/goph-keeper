package secret

import (
	"database/sql"
	"testing"
	"time"
)

func TestSecret_ToResponse(t *testing.T) {
	now := time.Now()
	secret := &Secret{
		ID:         "test-id",
		UserID:     1,
		Login:      "test-login",
		Password:   "test-password",
		Metadata:   map[string]interface{}{"key": "value"},
		BinaryData: []byte("binary data"),
		Version:    1,
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  sql.NullTime{Valid: false},
	}

	resp := secret.ToResponse()

	if resp.ID != secret.ID {
		t.Errorf("Expected ID %s, got %s", secret.ID, resp.ID)
	}
	if resp.Login != secret.Login {
		t.Errorf("Expected login %s, got %s", secret.Login, resp.Login)
	}
	if resp.Password != secret.Password {
		t.Errorf("Expected password %s, got %s", secret.Password, resp.Password)
	}
	if resp.Version != secret.Version {
		t.Errorf("Expected version %d, got %d", secret.Version, resp.Version)
	}
	if len(resp.BinaryData) != len(secret.BinaryData) {
		t.Errorf("Expected binary data length %d, got %d", len(secret.BinaryData), len(resp.BinaryData))
	}
	if resp.DeletedAt != nil {
		t.Error("Expected DeletedAt to be nil")
	}
}

func TestSecret_ToResponse_WithDeletedAt(t *testing.T) {
	now := time.Now()
	deletedTime := now.Add(-1 * time.Hour)
	secret := &Secret{
		ID:        "test-id",
		UserID:    1,
		Login:     "test-login",
		Password:  "test-password",
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: sql.NullTime{Valid: true, Time: deletedTime},
	}

	resp := secret.ToResponse()

	if resp.DeletedAt == nil {
		t.Fatal("Expected DeletedAt to be set")
	}
	if !resp.DeletedAt.Equal(deletedTime) {
		t.Errorf("Expected DeletedAt %v, got %v", deletedTime, *resp.DeletedAt)
	}
}

func TestSecret_ToResponseForSync(t *testing.T) {
	now := time.Now()
	smallData := []byte("small data")
	secret := &Secret{
		ID:         "test-id",
		UserID:     1,
		Login:      "test-login",
		Password:   "test-password",
		Metadata:   map[string]interface{}{"key": "value"},
		BinaryData: smallData,
		Version:    1,
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  sql.NullTime{Valid: false},
	}

	resp := secret.ToResponseForSync()

	if resp.ID != secret.ID {
		t.Errorf("Expected ID %s, got %s", secret.ID, resp.ID)
	}
	if len(resp.BinaryData) != len(smallData) {
		t.Errorf("Expected binary data length %d, got %d", len(smallData), len(resp.BinaryData))
	}
	if resp.BinaryDataSize != nil {
		t.Error("Expected BinaryDataSize to be nil for small data")
	}
}

func TestSecret_ToResponseForSync_LargeData(t *testing.T) {
	now := time.Now()
	// Create data larger than MinSizeForChunks (1 MB)
	largeData := make([]byte, MinSizeForChunks+1000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	secret := &Secret{
		ID:         "test-id",
		UserID:     1,
		Login:      "test-login",
		Password:   "test-password",
		BinaryData: largeData,
		Version:    1,
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  sql.NullTime{Valid: false},
	}

	resp := secret.ToResponseForSync()

	if resp.BinaryData != nil {
		t.Error("Expected BinaryData to be nil for large data")
	}
	if resp.BinaryDataSize == nil {
		t.Fatal("Expected BinaryDataSize to be set for large data")
	}
	expectedSize := int64(len(largeData))
	if *resp.BinaryDataSize != expectedSize {
		t.Errorf("Expected BinaryDataSize %d, got %d", expectedSize, *resp.BinaryDataSize)
	}
}

func TestSecret_ToResponseForSync_WithDeletedAt(t *testing.T) {
	now := time.Now()
	deletedTime := now.Add(-1 * time.Hour)
	secret := &Secret{
		ID:         "test-id",
		UserID:     1,
		Login:      "test-login",
		Password:   "test-password",
		BinaryData: []byte("data"),
		Version:    1,
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  sql.NullTime{Valid: true, Time: deletedTime},
	}

	resp := secret.ToResponseForSync()

	if resp.DeletedAt == nil {
		t.Fatal("Expected DeletedAt to be set")
	}
	if !resp.DeletedAt.Equal(deletedTime) {
		t.Errorf("Expected DeletedAt %v, got %v", deletedTime, *resp.DeletedAt)
	}
}

func TestSecret_ToResponseForSync_ExactlyAtThreshold(t *testing.T) {
	now := time.Now()
	// Create data exactly at MinSizeForChunks
	thresholdData := make([]byte, MinSizeForChunks)
	for i := range thresholdData {
		thresholdData[i] = byte(i % 256)
	}

	secret := &Secret{
		ID:         "test-id",
		UserID:     1,
		Login:      "test-login",
		Password:   "test-password",
		BinaryData: thresholdData,
		Version:    1,
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  sql.NullTime{Valid: false},
	}

	resp := secret.ToResponseForSync()

	// At exactly MinSizeForChunks, should NOT use chunked mode (only > MinSizeForChunks)
	if resp.BinaryData == nil || len(resp.BinaryData) != MinSizeForChunks {
		t.Error("Expected BinaryData to be included at exact threshold")
	}
	if resp.BinaryDataSize != nil {
		t.Error("Expected BinaryDataSize to be nil at exact threshold")
	}
}
