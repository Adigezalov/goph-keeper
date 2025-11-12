import { TAction, TSyncStatus } from '@shared/db'

export type TSecret = {
	localId: string
	id?: string
	login: string
	password: string // зашифрованный
	metadata: string
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
	metadata: string
}
