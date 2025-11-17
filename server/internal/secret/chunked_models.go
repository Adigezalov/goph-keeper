package secret

import (
	"time"
)

// ChunkedUploadSession представляет сессию загрузки большого файла по частям
type ChunkedUploadSession struct {
	UploadID    string    `json:"upload_id"`
	SecretID    string    `json:"secret_id"`
	UserID      string    `json:"user_id"`
	TotalChunks int       `json:"total_chunks"`
	TotalSize   int64     `json:"total_size"`
	Chunks      [][]byte  `json:"-"` // Временное хранение чанков
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// InitChunkedUploadRequest - запрос на инициализацию chunked upload
type InitChunkedUploadRequest struct {
	TotalChunks int               `json:"totalChunks"`
	TotalSize   int64             `json:"totalSize"`
	Metadata    map[string]string `json:"metadata"`
}

// InitChunkedUploadResponse - ответ на инициализацию
type InitChunkedUploadResponse struct {
	UploadID string `json:"uploadId"`
	SecretID string `json:"secretId"`
}

// UploadChunkRequest - запрос на загрузку одного чанка
type UploadChunkRequest struct {
	UploadID    string `json:"uploadId"`
	ChunkIndex  int    `json:"chunkIndex"`
	TotalChunks int    `json:"totalChunks"`
	Data        string `json:"data"` // base64
}

// UploadChunkResponse - ответ на загрузку чанка
type UploadChunkResponse struct {
	ChunkIndex int  `json:"chunkIndex"`
	Received   bool `json:"received"`
}

// FinalizeChunkedUploadRequest - запрос на завершение chunked upload
type FinalizeChunkedUploadRequest struct {
	UploadID string            `json:"uploadId"`
	Login    string            `json:"login"`
	Password string            `json:"password"`
	Metadata map[string]string `json:"metadata"`
	Version  *int              `json:"version,omitempty"`
}

// DownloadChunkResponse - ответ с одним чанком при скачивании
type DownloadChunkResponse struct {
	ChunkIndex  int    `json:"chunkIndex"`
	Data        string `json:"data"` // base64
	TotalChunks int    `json:"totalChunks"`
}
