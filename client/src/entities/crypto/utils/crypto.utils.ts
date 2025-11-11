import { LOCAL_CRYPTO_KEY_NAME } from '../constants'

// Генерирует новый CryptoKey и возвращает его в виде строки JWK
export const generateCryptoKey = async (): Promise<string> => {
	const key = await window.crypto.subtle.generateKey(
		{
			name: 'AES-GCM',
			length: 256,
		},
		true,
		['encrypt', 'decrypt'],
	)

	return await exportCryptoKey(key)
}

// Экспортирует CryptoKey в формат JWK и преобразует в строку для хранения
export const exportCryptoKey = async (key: CryptoKey): Promise<string> => {
	const exported = await window.crypto.subtle.exportKey('jwk', key)
	return JSON.stringify(exported)
}

// Импортирует CryptoKey из сохраненной строки JWK
export const importCryptoKey = async (keyString: string): Promise<CryptoKey> => {
	const keyData = JSON.parse(keyString)
	return await window.crypto.subtle.importKey(
		'jwk',
		keyData,
		{ name: 'AES-GCM', length: 256 },
		true,
		['encrypt', 'decrypt'],
	)
}

// Получает сохраненный ключ из localStorage
export const getCryptoKey = (): string | null => {
	return localStorage.getItem(LOCAL_CRYPTO_KEY_NAME)
}

// Удаляет ключ из localStorage
export const removeCryptoKey = () => {
	localStorage.removeItem(LOCAL_CRYPTO_KEY_NAME)
}

// Сохраняет ключ в localStorage
export const setCryptoKey = (key: string) => {
	localStorage.setItem(LOCAL_CRYPTO_KEY_NAME, key)
}
