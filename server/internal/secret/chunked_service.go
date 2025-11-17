package secret

import (
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ChunkedUploadService управляет сессиями chunked upload
type ChunkedUploadService struct {
	sessions map[string]*ChunkedUploadSession
	mu       sync.RWMutex
}

// NewChunkedUploadService создает новый сервис
func NewChunkedUploadService() *ChunkedUploadService {
	service := &ChunkedUploadService{
		sessions: make(map[string]*ChunkedUploadSession),
	}

	// Запускаем очистку просроченных сессий каждые 5 минут
	go service.cleanupExpiredSessions()

	return service
}

// InitUpload инициализирует новую сессию загрузки
func (s *ChunkedUploadService) InitUpload(userID string, totalChunks int, totalSize int64) (*ChunkedUploadSession, error) {
	uploadID := uuid.New().String()
	secretID := uuid.New().String()

	fmt.Printf("[ChunkedUpload] InitUpload - userID: %s, totalChunks: %d, totalSize: %d\n", userID, totalChunks, totalSize)

	session := &ChunkedUploadSession{
		UploadID:    uploadID,
		SecretID:    secretID,
		UserID:      userID,
		TotalChunks: totalChunks,
		TotalSize:   totalSize,
		Chunks:      make([][]byte, totalChunks),
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(30 * time.Minute), // Сессия живет 30 минут
	}

	s.mu.Lock()
	s.sessions[uploadID] = session
	s.mu.Unlock()

	return session, nil
}

// UploadChunk загружает один чанк
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

	// Декодируем base64
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("failed to decode chunk data: %w", err)
	}

	session.Chunks[chunkIndex] = decoded
	return nil
}

// GetCompleteData собирает все чанки и возвращает полные данные
func (s *ChunkedUploadService) GetCompleteData(uploadID string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[uploadID]
	if !exists {
		return nil, fmt.Errorf("upload session not found")
	}

	// Проверяем, что все чанки загружены
	for i, chunk := range session.Chunks {
		if chunk == nil {
			return nil, fmt.Errorf("chunk %d is missing", i)
		}
	}

	// Собираем все чанки в один массив
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

// CleanupSession удаляет сессию после завершения
func (s *ChunkedUploadService) CleanupSession(uploadID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, uploadID)
}

// GetSession возвращает сессию по ID
func (s *ChunkedUploadService) GetSession(uploadID string) (*ChunkedUploadSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[uploadID]
	if !exists {
		return nil, fmt.Errorf("upload session not found")
	}

	return session, nil
}

// cleanupExpiredSessions удаляет просроченные сессии
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

// SplitIntoChunks разбивает данные на чанки для отправки клиенту
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
