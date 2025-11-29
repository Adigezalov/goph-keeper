import { api } from '@shared/api'
import { IResponse } from '@shared/types'

import { SECRETS_URL } from '../constants/api.constants'
import {
	TCreateSecretRequest,
	TSecretResponse,
	TSyncResponse,
	TUpdateSecretRequest,
} from '../types/api.types'

export const createSecretApi = (
	data: TCreateSecretRequest,
): Promise<IResponse<TSecretResponse>> => {
	return api.post(SECRETS_URL.BASE, data)
}

export const getSecretsApi = (): Promise<IResponse<TSecretResponse[]>> => {
	return api.get(SECRETS_URL.BASE)
}

export const getSecretApi = (id: string): Promise<IResponse<TSecretResponse>> => {
	return api.get(SECRETS_URL.BY_ID(id))
}

export const updateSecretApi = (
	id: string,
	data: TUpdateSecretRequest,
): Promise<IResponse<TSecretResponse>> => {
	return api.put(SECRETS_URL.BY_ID(id), data)
}

export const deleteSecretApi = (id: string): Promise<IResponse<null>> => {
	return api.delete(SECRETS_URL.BY_ID(id))
}

export const syncSecretsApi = (since?: string): Promise<IResponse<TSyncResponse>> => {
	const url = since
		? `${SECRETS_URL.SYNC}?since=${encodeURIComponent(since)}`
		: SECRETS_URL.SYNC
	return api.get(url)
}
