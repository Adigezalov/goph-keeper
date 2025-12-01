package secret

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitChunkedUploadRequest(t *testing.T) {
	req := InitChunkedUploadRequest{
		TotalChunks: 10,
		TotalSize:   1024000,
	}

	assert.Equal(t, 10, req.TotalChunks)
	assert.Equal(t, int64(1024000), req.TotalSize)
}

func TestInitChunkedUploadResponse(t *testing.T) {
	resp := InitChunkedUploadResponse{
		UploadID: "upload-123",
		SecretID: "secret-456",
	}

	assert.Equal(t, "upload-123", resp.UploadID)
	assert.Equal(t, "secret-456", resp.SecretID)
}

func TestUploadChunkRequest(t *testing.T) {
	req := UploadChunkRequest{
		UploadID:   "upload-123",
		ChunkIndex: 5,
		Data:       "base64data==",
	}

	assert.Equal(t, "upload-123", req.UploadID)
	assert.Equal(t, 5, req.ChunkIndex)
	assert.Equal(t, "base64data==", req.Data)
}

func TestUploadChunkResponse(t *testing.T) {
	resp := UploadChunkResponse{
		ChunkIndex: 3,
		Received:   true,
	}

	assert.Equal(t, 3, resp.ChunkIndex)
	assert.True(t, resp.Received)
}

func TestFinalizeChunkedUploadRequest(t *testing.T) {
	version := 5
	req := FinalizeChunkedUploadRequest{
		UploadID: "upload-789",
		Login:    "user",
		Password: "pass",
		Metadata: map[string]string{
			"key": "value",
		},
		Version: &version,
	}

	assert.Equal(t, "upload-789", req.UploadID)
	assert.Equal(t, "user", req.Login)
	assert.Equal(t, "pass", req.Password)
	assert.NotNil(t, req.Version)
	assert.Equal(t, 5, *req.Version)
	assert.Equal(t, "value", req.Metadata["key"])
}

func TestFinalizeChunkedUploadRequest_WithoutVersion(t *testing.T) {
	req := FinalizeChunkedUploadRequest{
		UploadID: "upload-999",
		Login:    "user",
		Password: "pass",
		Version:  nil,
	}

	assert.Equal(t, "upload-999", req.UploadID)
	assert.Nil(t, req.Version)
}

func TestDownloadChunkResponse(t *testing.T) {
	resp := DownloadChunkResponse{
		ChunkIndex:  2,
		Data:        "chunkdata==",
		TotalChunks: 10,
	}

	assert.Equal(t, 2, resp.ChunkIndex)
	assert.Equal(t, "chunkdata==", resp.Data)
	assert.Equal(t, 10, resp.TotalChunks)
}

func TestChunkedUploadSession(t *testing.T) {
	now := time.Now()
	session := ChunkedUploadSession{
		UploadID:    "upload-123",
		SecretID:    "secret-456",
		UserID:      "user-789",
		TotalChunks: 5,
		TotalSize:   500000,
		Chunks:      make([][]byte, 5),
		CreatedAt:   now,
		ExpiresAt:   now.Add(1 * time.Hour),
	}

	assert.Equal(t, "upload-123", session.UploadID)
	assert.Equal(t, "secret-456", session.SecretID)
	assert.Equal(t, "user-789", session.UserID)
	assert.Equal(t, 5, session.TotalChunks)
	assert.Equal(t, int64(500000), session.TotalSize)
	assert.Len(t, session.Chunks, 5)
	assert.Equal(t, now, session.CreatedAt)
	assert.True(t, session.ExpiresAt.After(now))
}

func TestChunkedUploadSession_WithChunks(t *testing.T) {
	session := ChunkedUploadSession{
		UploadID:    "upload-999",
		SecretID:    "secret-999",
		UserID:      "user-999",
		TotalChunks: 3,
		TotalSize:   300,
		Chunks:      [][]byte{[]byte("chunk1"), []byte("chunk2"), []byte("chunk3")},
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	}

	assert.Len(t, session.Chunks, 3)
	assert.Equal(t, []byte("chunk1"), session.Chunks[0])
	assert.Equal(t, []byte("chunk2"), session.Chunks[1])
	assert.Equal(t, []byte("chunk3"), session.Chunks[2])
}
