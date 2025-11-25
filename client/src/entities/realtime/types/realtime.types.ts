export type SecretEventType = 'secret_created' | 'secret_updated' | 'secret_deleted'

export interface SecretEventMessage {
	type: SecretEventType
	secret_id: string
	user_id: number
	timestamp: string
}

export type RealtimeConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'reconnecting'

