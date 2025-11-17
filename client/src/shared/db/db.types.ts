// Статус синхронизации секрета
export type TSyncStatus = 'synced' | 'pending' | 'conflict' | 'deleted'
export type TAction = 'create' | 'update' | 'delete'

// Метаданные синхронизации (хранятся в IndexedDB)
export type TSyncMeta = {
	key: string // Ключ метаданных (например, 'lastSyncTime')
	value: string // Значение (ISO timestamp)
}