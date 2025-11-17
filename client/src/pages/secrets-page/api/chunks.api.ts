import { api } from '@shared/api'
import { IResponse } from '@shared/types'

import { SECRETS_URL } from '../constants/api.constants'
import { TSecretResponse } from '../types'

// Типы для chunked upload
export type TInitChunkedUploadRequest = {
	totalChunks: number
	totalSize: number
	metadata?: Record<string, string>
}

export type TInitChunkedUploadResponse = {
	uploadId: string
	secretId: string
}

export type TUploadChunkRequest = {
	uploadId: string
	chunkIndex: number
	totalChunks: number
	data: string // base64
}

export type TFinalizeChunkedUploadRequest = {
	uploadId: string
	login: string // зашифрованный
	password: string // зашифрованный
	metadata?: Record<string, string>
	version?: number // для update
}

/**
 * Инициализация chunked upload
 */
export const initChunkedUploadApi = (
	data: TInitChunkedUploadRequest,
): Promise<IResponse<TInitChunkedUploadResponse>> => {
	return api.post(`${SECRETS_URL.BASE}/chunks/init`, data)
}

/**
 * Загрузка одного чанка
 */
export const uploadChunkApi = (
	secretId: string,
	data: TUploadChunkRequest,
): Promise<IResponse<{ chunkIndex: number; received: boolean }>> => {
	return api.post(`${SECRETS_URL.BASE}/${secretId}/chunks`, data)
}

/**
 * Завершение chunked upload (создание секрета)
 */
export const finalizeChunkedUploadApi = (
	secretId: string,
	data: TFinalizeChunkedUploadRequest,
): Promise<IResponse<TSecretResponse>> => {
	return api.post(`${SECRETS_URL.BASE}/${secretId}/chunks/finalize`, data)
}

/**
 * Получение чанка при скачивании
 */
export const downloadChunkApi = (
	secretId: string,
	chunkIndex: number,
): Promise<IResponse<{ chunkIndex: number; data: string; totalChunks: number }>> => {
	return api.get(`${SECRETS_URL.BASE}/${secretId}/chunks/${chunkIndex}`)
}

