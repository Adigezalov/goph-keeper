import axios, { AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { StatusCodes } from 'http-status-codes'
import i18next from 'i18next'

import { showToastNotification } from '@shared/toast-notification'
import { TOAST_SEVERITY } from '@shared/uikit/toast'

import { REFRESH } from '../constants'
import { getAccessToken, removeTokens, setTokens } from '../tokens'
import { getRealtimeSessionID } from '../realtime'

let isRefreshing = false
let failedQueue: Array<{
	resolve: (value?: any) => void
	reject: (reason?: any) => void
}> = []

const processQueue = (error: unknown, token: string | null = null) => {
	failedQueue.forEach(({ resolve, reject }) => {
		if (error) {
			reject(error)
		} else {
			resolve(token)
		}
	})

	failedQueue = []
}

export const api = axios.create({
	withCredentials: true,
})

api.interceptors.request.use(
	(config: InternalAxiosRequestConfig): InternalAxiosRequestConfig => {
		const token = getAccessToken()

		config.headers['Authorization'] = `Bearer ${token}`
		config.headers['Content-Type'] = 'application/json'

		const sessionID = getRealtimeSessionID()
		if (sessionID) {
			config.headers['X-Session-ID'] = sessionID
		}

		return config
	},
)

api.interceptors.response.use(
	(response: AxiosResponse): AxiosResponse => {
		if (response.data?.access_token) {
			setTokens(response.data.access_token)
		}
		return response
	},

	async (error) => {
		const originalRequest = error.config

		if (!error.response) {
			return Promise.reject(error)
		}

		// Не показываем toast для конфликтов версий (409) - они обрабатываются отдельно
		if (error.response.status !== StatusCodes.CONFLICT) {
			showToastNotification({
				message: error?.response?.data
					? error?.response?.data
					: i18next.t('error.unexpected_error'),
				header: i18next.t('error.error'),
				severity: TOAST_SEVERITY.ERROR,
			})
		}

		if (error.response.status === StatusCodes.UNAUTHORIZED && !error.config._isRetry) {
			if (isRefreshing) {
				return new Promise((resolve, reject) => {
					failedQueue.push({ resolve, reject })
				})
					.then((token) => {
						originalRequest.headers['Authorization'] = `Bearer ${token}`
						return api.request(originalRequest)
					})
					.catch((err) => {
						return Promise.reject(err)
					})
			}

			originalRequest._isRetry = true
			isRefreshing = true

			try {
				const response: any = await axios({
					method: 'GET',
					url: REFRESH,
					withCredentials: true,
				})

				const accessToken = response.data.access_token

				setTokens(accessToken)

				processQueue(null, accessToken)

				originalRequest.headers['Authorization'] = `Bearer ${accessToken}`
				return api.request(originalRequest)
			} catch (refreshError) {
				processQueue(refreshError, null)

				removeTokens()
				window.location.reload()
				return Promise.reject(refreshError)
			} finally {
				isRefreshing = false
			}
		}

		if (error.response.status !== StatusCodes.CONFLICT) {
			showToastNotification({
				message: error?.response?.data
					? error?.response?.data
					: i18next.t('error.unexpected_error'),
				header: i18next.t('error.error'),
				severity: TOAST_SEVERITY.ERROR,
			})
		}

		return Promise.reject(error)
	},
)
