// Утилита для получения sessionID из localStorage
// Используется в HTTP interceptor для передачи в заголовках
export const getRealtimeSessionID = (): string | null => {
	return localStorage.getItem('realtime_session_id')
}

