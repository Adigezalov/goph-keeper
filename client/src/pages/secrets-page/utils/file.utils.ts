export const extractFileData = async (file: File) => {
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

