package secret

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunkedUploadService_InitUpload(t *testing.T) {
	service := NewChunkedUploadService()

	session, err := service.InitUpload("user-1", 5, 1000)
	require.NoError(t, err)
	assert.NotEmpty(t, session.UploadID)
	assert.NotEmpty(t, session.SecretID)
	assert.Equal(t, "user-1", session.UserID)
	assert.Equal(t, 5, session.TotalChunks)
	assert.Equal(t, int64(1000), session.TotalSize)
	assert.Equal(t, 5, len(session.Chunks))
}

func TestChunkedUploadService_UploadChunk(t *testing.T) {
	service := NewChunkedUploadService()

	session, err := service.InitUpload("user-1", 3, 300)
	require.NoError(t, err)

	data := []byte("test data")
	encoded := base64.StdEncoding.EncodeToString(data)

	err = service.UploadChunk(session.UploadID, 0, encoded)
	assert.NoError(t, err)

	retrievedSession, err := service.GetSession(session.UploadID)
	require.NoError(t, err)
	assert.NotNil(t, retrievedSession.Chunks[0])
}

func TestChunkedUploadService_UploadChunk_InvalidIndex(t *testing.T) {
	service := NewChunkedUploadService()

	session, err := service.InitUpload("user-1", 3, 300)
	require.NoError(t, err)

	data := []byte("test data")
	encoded := base64.StdEncoding.EncodeToString(data)

	err = service.UploadChunk(session.UploadID, 5, encoded)
	assert.Error(t, err)
}

func TestChunkedUploadService_UploadChunk_SessionNotFound(t *testing.T) {
	service := NewChunkedUploadService()

	data := []byte("test data")
	encoded := base64.StdEncoding.EncodeToString(data)

	err := service.UploadChunk("non-existent-id", 0, encoded)
	assert.Error(t, err)
}

func TestChunkedUploadService_GetCompleteData(t *testing.T) {
	service := NewChunkedUploadService()

	session, err := service.InitUpload("user-1", 3, 300)
	require.NoError(t, err)

	chunk1 := []byte("chunk1")
	chunk2 := []byte("chunk2")
	chunk3 := []byte("chunk3")

	err = service.UploadChunk(session.UploadID, 0, base64.StdEncoding.EncodeToString(chunk1))
	require.NoError(t, err)
	err = service.UploadChunk(session.UploadID, 1, base64.StdEncoding.EncodeToString(chunk2))
	require.NoError(t, err)
	err = service.UploadChunk(session.UploadID, 2, base64.StdEncoding.EncodeToString(chunk3))
	require.NoError(t, err)

	completeData, err := service.GetCompleteData(session.UploadID)
	require.NoError(t, err)

	expected := append(append(chunk1, chunk2...), chunk3...)
	assert.Equal(t, expected, completeData)
}

func TestChunkedUploadService_GetCompleteData_MissingChunk(t *testing.T) {
	service := NewChunkedUploadService()

	session, err := service.InitUpload("user-1", 3, 300)
	require.NoError(t, err)

	chunk1 := []byte("chunk1")
	err = service.UploadChunk(session.UploadID, 0, base64.StdEncoding.EncodeToString(chunk1))
	require.NoError(t, err)

	_, err = service.GetCompleteData(session.UploadID)
	assert.Error(t, err)
}

func TestChunkedUploadService_GetCompleteData_SessionNotFound(t *testing.T) {
	service := NewChunkedUploadService()

	_, err := service.GetCompleteData("non-existent-id")
	assert.Error(t, err)
}

func TestChunkedUploadService_CleanupSession(t *testing.T) {
	service := NewChunkedUploadService()

	session, err := service.InitUpload("user-1", 3, 300)
	require.NoError(t, err)

	_, err = service.GetSession(session.UploadID)
	assert.NoError(t, err)

	service.CleanupSession(session.UploadID)

	_, err = service.GetSession(session.UploadID)
	assert.Error(t, err)
}

func TestChunkedUploadService_GetSession(t *testing.T) {
	service := NewChunkedUploadService()

	session, err := service.InitUpload("user-1", 3, 300)
	require.NoError(t, err)

	retrievedSession, err := service.GetSession(session.UploadID)
	require.NoError(t, err)
	assert.Equal(t, session.UploadID, retrievedSession.UploadID)
	assert.Equal(t, session.SecretID, retrievedSession.SecretID)
	assert.Equal(t, session.UserID, retrievedSession.UserID)
}

func TestChunkedUploadService_GetSession_NotFound(t *testing.T) {
	service := NewChunkedUploadService()

	_, err := service.GetSession("non-existent-id")
	assert.Error(t, err)
}

func TestChunkedUploadService_CleanupExpiredSessions(t *testing.T) {
	service := NewChunkedUploadService()

	expiredSession := &ChunkedUploadSession{
		UploadID:    "expired-upload-id",
		SecretID:    "test-secret-id",
		UserID:      "user-1",
		TotalChunks: 1,
		TotalSize:   100,
		Chunks:      make([][]byte, 1),
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(-1 * time.Minute),
	}

	service.mu.Lock()
	service.sessions[expiredSession.UploadID] = expiredSession
	service.mu.Unlock()

	service.CleanupSession(expiredSession.UploadID)

	_, err := service.GetSession(expiredSession.UploadID)
	assert.Error(t, err)
}

func TestSplitIntoChunks(t *testing.T) {
	data := []byte("12345678901234567890")
	chunkSize := 5

	chunks := SplitIntoChunks(data, chunkSize)

	assert.Equal(t, 4, len(chunks))
	assert.Equal(t, []byte("12345"), chunks[0])
	assert.Equal(t, []byte("67890"), chunks[1])
	assert.Equal(t, []byte("12345"), chunks[2])
	assert.Equal(t, []byte("67890"), chunks[3])
}

func TestSplitIntoChunks_ExactSize(t *testing.T) {
	data := []byte("12345")
	chunkSize := 5

	chunks := SplitIntoChunks(data, chunkSize)

	assert.Equal(t, 1, len(chunks))
	assert.Equal(t, data, chunks[0])
}

func TestSplitIntoChunks_SmallerThanChunkSize(t *testing.T) {
	data := []byte("123")
	chunkSize := 5

	chunks := SplitIntoChunks(data, chunkSize)

	assert.Equal(t, 1, len(chunks))
	assert.Equal(t, data, chunks[0])
}
