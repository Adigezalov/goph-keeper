import { LOCAL_CRYPTO_KEY_NAME } from '../constants'

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

export const exportCryptoKey = async (key: CryptoKey): Promise<string> => {
	const exported = await window.crypto.subtle.exportKey('jwk', key)
	return JSON.stringify(exported)
}

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

export const getCryptoKey = (): string | null => {
	return localStorage.getItem(LOCAL_CRYPTO_KEY_NAME)
}

export const removeCryptoKey = () => {
	localStorage.removeItem(LOCAL_CRYPTO_KEY_NAME)
}

export const setCryptoKey = (key: string) => {
	localStorage.setItem(LOCAL_CRYPTO_KEY_NAME, key)
}
