export const CHUNK_SIZE = 100 * 1024

export const MIN_SIZE_FOR_CHUNKS = 1 * 1024 * 1024

export const splitIntoChunks = (data: Uint8Array, chunkSize: number = CHUNK_SIZE): Uint8Array[] => {
	const chunks: Uint8Array[] = []
	
	for (let offset = 0; offset < data.length; offset += chunkSize) {
		const chunk = data.slice(offset, Math.min(offset + chunkSize, data.length))
		chunks.push(chunk)
	}
	
	return chunks
}

export const mergeChunks = (chunks: Uint8Array[]): Uint8Array => {
	const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0)
	const result = new Uint8Array(totalLength)
	
	let offset = 0
	for (const chunk of chunks) {
		result.set(chunk, offset)
		offset += chunk.length
	}
	
	return result
}

export const calculateChunksCount = (dataSize: number, chunkSize: number = CHUNK_SIZE): number => {
	return Math.ceil(dataSize / chunkSize)
}

export const shouldUseChunks = (dataSize: number): boolean => {
	return dataSize > MIN_SIZE_FOR_CHUNKS
}

