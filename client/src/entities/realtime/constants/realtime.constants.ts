import { BASE_APP_URL } from '@shared/constants'

// WebSocket URL для realtime уведомлений
export const REALTIME_WS_URL = (token: string, sessionID?: string) => {
	const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
	const host = window.location.host
	const url = `${protocol}//${host}${BASE_APP_URL}/v1/realtime?token=${encodeURIComponent(token)}`
	
	if (sessionID) {
		return `${url}&session_id=${encodeURIComponent(sessionID)}`
	}
	
	return url
}

// Debounce время для синхронизации при получении событий (мс)
export const SYNC_DEBOUNCE_MS = 500

// Максимальное количество попыток переподключения
export const MAX_RECONNECT_ATTEMPTS = 10

// Базовая задержка перед переподключением (мс)
export const RECONNECT_BASE_DELAY_MS = 1000

// Максимальная задержка перед переподключением (мс)
export const RECONNECT_MAX_DELAY_MS = 30000

