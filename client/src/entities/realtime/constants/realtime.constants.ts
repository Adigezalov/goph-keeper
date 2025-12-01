import { BASE_APP_URL } from '@shared/constants'

export const REALTIME_WS_URL = (token: string, sessionID?: string) => {
	const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
	const host = window.location.host
	const url = `${protocol}//${host}${BASE_APP_URL}/v1/realtime?token=${encodeURIComponent(token)}`
	
	if (sessionID) {
		return `${url}&session_id=${encodeURIComponent(sessionID)}`
	}
	
	return url
}

export const SYNC_DEBOUNCE_MS = 500

export const MAX_RECONNECT_ATTEMPTS = 10

export const RECONNECT_BASE_DELAY_MS = 1000

export const RECONNECT_MAX_DELAY_MS = 30000

