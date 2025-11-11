import { api } from '@shared/api'
import { IResponse } from '@shared/types'

import { AUTH_URL } from '../constants'
import { TAuth } from '../types'

export const loginApi = (data: TAuth): Promise<IResponse<{ access_token: string }>> => {
	return api.post(AUTH_URL.LOGIN, data)
}

export const registrationApi = (
	data: TAuth,
): Promise<IResponse<{ access_token: string }>> => {
	return api.post(AUTH_URL.REGISTRATION, data)
}

export const refreshApi = (): Promise<IResponse<{ access_token: string }>> => {
	return api.get(AUTH_URL.REFRESH)
}

export const logoutApi = (): Promise<IResponse<null>> => {
	return api.get(AUTH_URL.LOGOUT)
}

export const logoutAllApi = (): Promise<IResponse<null>> => {
	return api.get(AUTH_URL.LOGOUT_ALL)
}
