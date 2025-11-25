export type TSecretResponse = {
	id: string
	login: string
	password: string
	metadata?: Record<string, string>
	binary_data?: string
	binary_data_size?: number
	version: number
	created_at: string
	updated_at: string
	deleted_at?: string
}

export type TCreateSecretRequest = {
	login: string
	password: string
	metadata?: Record<string, string>
	binary_data?: string
}

export type TUpdateSecretRequest = {
	login: string
	password: string
	metadata?: Record<string, string>
	binary_data?: string
	version: number
}

export type TSyncResponse = {
	secrets: TSecretResponse[]
	server_time: string
}

