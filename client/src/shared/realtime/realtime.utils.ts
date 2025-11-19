export const getRealtimeSessionID = (): string | null => {
	return localStorage.getItem('realtime_session_id')
}

