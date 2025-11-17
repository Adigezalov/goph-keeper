import { MAX_BINARY_DATA_SIZE } from '../constants/secret.constants'

export const extractFileData = async (file: File) => {
	// Проверяем размер файла
	if (file.size > MAX_BINARY_DATA_SIZE) {
		const maxSizeMB = MAX_BINARY_DATA_SIZE / (1024 * 1024)
		const fileSizeMB = (file.size / (1024 * 1024)).toFixed(2)
		throw new Error(
			`Файл слишком большой: ${fileSizeMB} МБ. Максимальный размер: ${maxSizeMB} МБ`,
		)
	}

	// Читаем файл как ArrayBuffer
	const arrayBuffer = await file.arrayBuffer()
	const binaryData = new Uint8Array(arrayBuffer)

	// Извлекаем имя файла и расширение
	const fileNameParts = file.name.split('.')
	const fileExtension = fileNameParts.length > 1 ? fileNameParts.pop() || '' : ''
	const fileName = fileNameParts.join('.')

	// Формируем metadata
	const metadata: Record<string, string> = {
		fileName,
		fileExtension,
		fileSize: String(file.size),
	}

	return { binaryData, metadata }
}

