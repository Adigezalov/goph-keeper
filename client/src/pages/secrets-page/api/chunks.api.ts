import { api } from '@shared/api'
import { IResponse } from '@shared/types'

import { SECRETS_URL } from '../constants/api.constants'
import { TSecretResponse } from '../types'

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
	data: string
}

export type TFinalizeChunkedUploadRequest = {
	uploadId: string
	login: string
	password: string
	metadata?: Record<string, string>
	version?: number
}

export const initChunkedUploadApi = (
	data: TInitChunkedUploadRequest,
): Promise<IResponse<TInitChunkedUploadResponse>> => {
	return api.post(`${SECRETS_URL.BASE}/chunks/init`, data)
}

export const uploadChunkApi = (
	secretId: string,
	data: TUploadChunkRequest,
): Promise<IResponse<{ chunkIndex: number; received: boolean }>> => {
	return api.post(`${SECRETS_URL.BASE}/${secretId}/chunks`, data)
}

export const finalizeChunkedUploadApi = (
	secretId: string,
	data: TFinalizeChunkedUploadRequest,
): Promise<IResponse<TSecretResponse>> => {
	return api.post(`${SECRETS_URL.BASE}/${secretId}/chunks/finalize`, data)
}

export const downloadChunkApi = (
	secretId: string,
	chunkIndex: number,
): Promise<IResponse<{ chunkIndex: number; data: string; totalChunks: number }>> => {
	return api.get(`${SECRETS_URL.BASE}/${secretId}/chunks/${chunkIndex}`)
}

