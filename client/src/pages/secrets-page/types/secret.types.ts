import { TAction, TSyncStatus } from '@shared/db'

import { TSecretResponse } from './api.types'

export type TSecret = {
	localId: string
	id?: string
	login: string
	password: string
	metadata?: Record<string, string>
	binaryData?: Uint8Array
	version: number
	syncStatus: TSyncStatus
	createdAt: number
	updatedAt: number
	deletedAt?: number
}

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

export type TConflictResolution = {
	secretId: string
	localId: string
	localVersion: TSecret
	serverVersion: TSecretResponse
	serverBinaryData?: Uint8Array
	conflictType: 'update' | 'sync'
	timestamp: number
}
