// Типы для API запросов/ответов с сервера

// Response типы (данные с сервера - зашифрованные)
export type TSecretResponse = {
	id: string // UUID с сервера
	login: string // зашифрованный base64
	password: string // зашифрованный base64
	metadata?: Record<string, string> // НЕ зашифрованный
	binary_data?: string // зашифрованный base64
	version: number
	created_at: string // ISO timestamp
	updated_at: string // ISO timestamp
	deleted_at?: string // ISO timestamp для soft delete
}

// Request для создания секрета (отправка на сервер)
export type TCreateSecretRequest = {
	login: string // зашифрованный base64
	password: string // зашифрованный base64
	metadata?: Record<string, string>
	binary_data?: string // зашифрованный base64
}

// Request для обновления секрета
export type TUpdateSecretRequest = {
	login: string // зашифрованный base64
	password: string // зашифрованный base64
	metadata?: Record<string, string>
	binary_data?: string // зашифрованный base64
	version: number // для оптимистической блокировки
}

// Response для синхронизации
export type TSyncResponse = {
	secrets: TSecretResponse[]
	server_time: string // ISO timestamp для следующего запроса
}

