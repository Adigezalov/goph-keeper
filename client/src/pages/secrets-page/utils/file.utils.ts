import i18next from 'i18next'

import { MAX_BINARY_DATA_SIZE } from '../constants/secret.constants'

export const extractFileData = async (file: File) => {
	if (file.size > MAX_BINARY_DATA_SIZE) {
		const maxSizeMB = MAX_BINARY_DATA_SIZE / (1024 * 1024)
		const fileSizeMB = (file.size / (1024 * 1024)).toFixed(2)
		throw new Error(
			i18next.t('file.too_large', {
				fileSize: fileSizeMB,
				maxSize: maxSizeMB,
			}),
		)
	}

	const arrayBuffer = await file.arrayBuffer()
	const binaryData = new Uint8Array(arrayBuffer)

	const fileNameParts = file.name.split('.')
	const fileExtension = fileNameParts.length > 1 ? fileNameParts.pop() || '' : ''
	const fileName = fileNameParts.join('.')

	const metadata: Record<string, string> = {
		fileName,
		fileExtension,
		fileSize: String(file.size),
	}

	return { binaryData, metadata }
}

