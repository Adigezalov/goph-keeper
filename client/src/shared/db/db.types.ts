export type TSyncStatus = 'synced' | 'pending' | 'conflict' | 'deleted'
export type TAction = 'create' | 'update' | 'delete'

export type TSyncMeta = {
	key: string
	value: string
}