import { TAction, TSyncStatus } from '@shared/db'

export type TSecret = {
	localId: string
	id?: string //id с сервера
	login: string
	password: string
	metadata?: Record<string, string> // fileName, fileExtension, fileSize, app и т.д.
	binaryData?: Uint8Array
	version: number
	syncStatus: TSyncStatus
	createdAt: number
	updatedAt: number
	deletedAt?: number
}

// Очередь синхронизации
export type TSyncQueue = {
	id: string
	action: TAction
	secretId: string
	data: TSecret
	timestamp: number
	retryCount: number
}

export type TSecretForSave = {
	login: string
	password: string
	metadata?: Record<string, string>
	binaryData?: Uint8Array
}
