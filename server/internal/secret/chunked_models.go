package secret

import (
	"time"
)

type ChunkedUploadSession struct {
	UploadID    string    `json:"upload_id"`
	SecretID    string    `json:"secret_id"`
	UserID      string    `json:"user_id"`
	TotalChunks int       `json:"total_chunks"`
	TotalSize   int64     `json:"total_size"`
	Chunks      [][]byte  `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type InitChunkedUploadRequest struct {
	TotalChunks int               `json:"totalChunks"`
	TotalSize   int64             `json:"totalSize"`
	Metadata    map[string]string `json:"metadata"`
}

type InitChunkedUploadResponse struct {
	UploadID string `json:"uploadId"`
	SecretID string `json:"secretId"`
}

type UploadChunkRequest struct {
	UploadID    string `json:"uploadId"`
	ChunkIndex  int    `json:"chunkIndex"`
	TotalChunks int    `json:"totalChunks"`
	Data        string `json:"data"`
}

type UploadChunkResponse struct {
	ChunkIndex int  `json:"chunkIndex"`
	Received   bool `json:"received"`
}

type FinalizeChunkedUploadRequest struct {
	UploadID string            `json:"uploadId"`
	Login    string            `json:"login"`
	Password string            `json:"password"`
	Metadata map[string]string `json:"metadata"`
	Version  *int              `json:"version,omitempty"`
}

type DownloadChunkResponse struct {
	ChunkIndex  int    `json:"chunkIndex"`
	Data        string `json:"data"`
	TotalChunks int    `json:"totalChunks"`
}
