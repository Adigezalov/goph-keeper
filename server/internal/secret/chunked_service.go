package secret

import (
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/logger"
	"github.com/google/uuid"
)

type ChunkedUploadService struct {
	sessions map[string]*ChunkedUploadSession
	mu       sync.RWMutex
}

func NewChunkedUploadService() *ChunkedUploadService {
	service := &ChunkedUploadService{
		sessions: make(map[string]*ChunkedUploadSession),
	}

	go service.cleanupExpiredSessions()

	return service
}

func (s *ChunkedUploadService) InitUpload(userID string, totalChunks int, totalSize int64) (*ChunkedUploadSession, error) {
	uploadID := uuid.New().String()
	secretID := uuid.New().String()

	logger.Log.WithFields(map[string]interface{}{
		"user_id":      userID,
		"upload_id":    uploadID,
		"secret_id":    secretID,
		"total_chunks": totalChunks,
		"total_size":   totalSize,
	}).Info("[ChunkedUpload] Инициализация загрузки")

	session := &ChunkedUploadSession{
		UploadID:    uploadID,
		SecretID:    secretID,
		UserID:      userID,
		TotalChunks: totalChunks,
		TotalSize:   totalSize,
		Chunks:      make([][]byte, totalChunks),
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	}

	s.mu.Lock()
	s.sessions[uploadID] = session
	s.mu.Unlock()

	return session, nil
}

func (s *ChunkedUploadService) UploadChunk(uploadID string, chunkIndex int, data string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[uploadID]
	if !exists {
		return fmt.Errorf("upload session not found or expired")
	}

	if chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		return fmt.Errorf("invalid chunk index: %d (totalChunks: %d)", chunkIndex, session.TotalChunks)
	}

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("failed to decode chunk data: %w", err)
	}

	session.Chunks[chunkIndex] = decoded
	return nil
}

func (s *ChunkedUploadService) GetCompleteData(uploadID string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[uploadID]
	if !exists {
		return nil, fmt.Errorf("upload session not found")
	}

	for i, chunk := range session.Chunks {
		if chunk == nil {
			return nil, fmt.Errorf("chunk %d is missing", i)
		}
	}

	totalSize := 0
	for _, chunk := range session.Chunks {
		totalSize += len(chunk)
	}

	result := make([]byte, 0, totalSize)
	for _, chunk := range session.Chunks {
		result = append(result, chunk...)
	}

	return result, nil
}

func (s *ChunkedUploadService) CleanupSession(uploadID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, uploadID)
}

func (s *ChunkedUploadService) GetSession(uploadID string) (*ChunkedUploadSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[uploadID]
	if !exists {
		return nil, fmt.Errorf("upload session not found")
	}

	return session, nil
}

func (s *ChunkedUploadService) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for id, session := range s.sessions {
			if now.After(session.ExpiresAt) {
				delete(s.sessions, id)
			}
		}
		s.mu.Unlock()
	}
}

func SplitIntoChunks(data []byte, chunkSize int) [][]byte {
	var chunks [][]byte

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}

	return chunks
}
